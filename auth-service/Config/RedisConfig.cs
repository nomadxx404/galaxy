namespace auth_service.Config
{
    public class RedisConfig
    {
        public string REDIS_HOST => Env.Get("REDIS_HOST");
        public int REDIS_PORT => Env.GetInt("REDIS_PORT");
        public string REDIS_PASSWORD => Env.Get("REDIS_PASSWORD");

        public string ConnectionString =>
            $"{REDIS_HOST}:{REDIS_PORT},password={REDIS_PASSWORD}";
    }
}
