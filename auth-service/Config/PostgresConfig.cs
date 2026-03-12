namespace auth_service.Config
{
    public class PostgresConfig
    {
        public string POSTGRES_HOST => Env.Get("POSTGRES_HOST");
        public int POSTGRES_PORT => Env.GetInt("POSTGRES_PORT");
        public string POSTGRES_DB => Env.Get("POSTGRES_DB");
        public string POSTGRES_USER => Env.Get("POSTGRES_USER");
        public string POSTGRES_PASSWORD => Env.Get("POSTGRES_PASSWORD");

        public string ConnectionString =>
            $"Host={POSTGRES_HOST};Port={POSTGRES_PORT};Database={POSTGRES_DB};Username={POSTGRES_USER};Password={POSTGRES_PASSWORD}";
    }
}
