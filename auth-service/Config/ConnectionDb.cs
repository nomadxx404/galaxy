using Npgsql;

namespace auth_service.Config
{
    public static class ConnectionDb
    {
        public static IServiceCollection ConnectionPostgres(this IServiceCollection services)
        {
            string RequiredEnv(string name) =>
                Environment.GetEnvironmentVariable(name)
                ?? throw new InvalidOperationException($"Missing env var: {name}");

            var csb = new NpgsqlConnectionStringBuilder
            {
                Host = RequiredEnv("POSTGRES_HOST"),
                Port = int.Parse(RequiredEnv("POSTGRES_PORT")),
                Database = RequiredEnv("POSTGRES_DB"),
                Username = RequiredEnv("POSTGRES_USER"),
                Password = RequiredEnv("POSTGRES_PASSWORD"),

                Pooling = true,
                Timeout = 5,
                CommandTimeout = 30,
            };

            services.AddSingleton(NpgsqlDataSource.Create(csb.ConnectionString));

            return services;
        }
    }
}
