using auth_service.Config;
using Npgsql;

namespace auth_service.Data
{
    public class PostgresConnection
    {
        private readonly string _conn;

        public PostgresConnection(AppConfig config)
        {
            _conn = config.PostgresConfig.ConnectionString;
        }

        public async Task<NpgsqlConnection> OpenConnectionAsync()
        {
            var connection = new NpgsqlConnection(_conn);
            await connection.OpenAsync();
            return connection;
        }
    }
}