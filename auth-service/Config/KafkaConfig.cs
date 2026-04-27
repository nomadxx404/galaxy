namespace auth_service.Config
{
    public class KafkaConfig
    {
        public string BROKERS => Env.Get("KAFKA_BOOTSTRAP_SERVERS");
        public string EVENTS_TOPIC => Env.Get("KAFKA_AUTH_EVENTS_TOPIC");
    }
}
