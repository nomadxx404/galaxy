using auth_service.DTO.Request;
using auth_service.DTO.Response;
using Npgsql;

namespace auth_service.Repositories.Interfaces
{
    public interface IAccountRepository
    {
        public Task<AccountResponse?> GetAccount(string account_uuid, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<AccountUpdateResponse?> UpdateProfile(string account_uuid, AccountUpdateRequest request, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<string> GetCurrentPassword(string account_uuid, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task ChangePassword(string account_uuid, string passwordHash, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task DeleteAccount(string account_uuid, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<IEnumerable<AccountBatchResponse>> GetAccountsByUuids(IEnumerable<string> account_uuids, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
    }
}
