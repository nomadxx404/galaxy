using auth_service.Config;
using StackExchange.Redis;

namespace auth_service.Data
{
    public class RedisConnection
    {
        private readonly ConnectionMultiplexer _connection;

        public RedisConnection(RedisConfig config)
        {
            _connection = ConnectionMultiplexer.Connect(config.ConnectionString);
        }

        public IDatabase Database => _connection.GetDatabase();
    }
}
