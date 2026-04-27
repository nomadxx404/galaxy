using auth_service.Config;
using auth_service.Data;
using auth_service.Repositories.Interfaces;
using Confluent.Kafka;

public class Producer : BackgroundService
{
    private readonly ILogger<Producer> _logger;
    private readonly IServiceProvider _serviceProvider; 
    private readonly IProducer<string, string> _producer;
    private readonly AppConfig _appConfig;

    public Producer(
        ILogger<Producer> logger,
        IServiceProvider serviceProvider, 
        AppConfig appConfig)
    {
        _logger = logger;
        _serviceProvider = serviceProvider;
        _appConfig = appConfig;

        var producerConfig = new ProducerConfig
        {
            BootstrapServers = _appConfig.KafkaConfig.BROKERS,
            Acks = Acks.All,
            EnableIdempotence = true,
            MessageTimeoutMs = 10000,
            MessageSendMaxRetries = 2,
        };

        _producer = new ProducerBuilder<string, string>(producerConfig).Build();
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        _logger.LogInformation("[KAFKA] Outbox Producer Worker started");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                using (var scope = _serviceProvider.CreateScope())
                {
                    var outboxRepo = scope.ServiceProvider.GetRequiredService<IOutboxEventsRepository>();
                    var conn = scope.ServiceProvider.GetRequiredService<PostgresConnection>();

                    await ProcessOutboxEvents(outboxRepo, conn, stoppingToken);
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[KAFKA] Critical error in Producer process");
            }

            await Task.Delay(1000, stoppingToken);
        }
    }

    private async Task ProcessOutboxEvents(
        IOutboxEventsRepository outboxRepo,
        PostgresConnection conn,
        CancellationToken ct)
    {
        using var connection = await conn.OpenConnectionAsync();
        var events = await outboxRepo.GetUnprocessedEvents(connection);

        if (events == null || !events.Any()) return;

        foreach (var ev in events)
        {
            try
            {
                var message = new Message<string, string>
                {
                    Key = ev.AccountUuid,
                    Value = ev.Payload
                };

                await _producer.ProduceAsync(_appConfig.KafkaConfig.EVENTS_TOPIC, message, ct);

                await outboxRepo.MarkEventProcessed(ev.MessageUuid, connection);
            }
            catch (KafkaException ex)
            {
                _logger.LogError(ex, "[KAFKA] Failed to send message {MessageUuid}", ev.MessageUuid);
                await outboxRepo.MarkEventFailed(ev.MessageUuid, ex.Message, connection);
            }
        }
    }

    public override void Dispose()
    {
        _producer.Flush(TimeSpan.FromSeconds(10));
        _producer.Dispose();
        base.Dispose();
    }
}