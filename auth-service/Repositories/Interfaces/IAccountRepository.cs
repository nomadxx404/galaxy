using auth_service.DTO.Entity;
using auth_service.DTO.Request;
using auth_service.DTO.Response;
using auth_service.Models.Request;
using Npgsql;

namespace auth_service.Repositories.Interfaces
{
    public interface IAccountRepository
    {
        public Task<AccountResponse?> GetAccount(string account_uuid, NpgsqlConnection conn);
        public Task<AccountUpdateResponse?> UpdateProfile(string account_uuid, AccountUpdateRequest request, NpgsqlConnection conn);
        public Task<string> GetCurrentPassword(string account_uuid, NpgsqlConnection conn);
        public Task ChangePassword(string account_uuid, string passwordHash, NpgsqlConnection conn);
        public Task DeleteAccount(string account_uuid, NpgsqlConnection conn);
        public Task<IEnumerable<AccountBatchResponse>> GetAccountsByUuids(IEnumerable<string> account_uuids, NpgsqlConnection conn);
    }
}
