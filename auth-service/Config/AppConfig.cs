namespace auth_service.Config
{
    public class AppConfig
    {
        public PostgresConfig PostgresConfig { get; } = new();
        public RedisConfig RedisConfig { get; } = new();
        public JwtConfig JwtConfig { get; } = new();
        public ArgonConfig ArgonConfig { get; } = new();
        public AuthConfig AuthConfig { get; } = new();
        public KafkaConfig KafkaConfig { get; } = new();
    }
}
