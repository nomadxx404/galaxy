using auth_service.DTOs.Entity;
using auth_service.Models.Request;
using Npgsql;

namespace auth_service.Repositories.Interfaces
{
    public interface IAuthRepository
    {
        public Task AccountCreate(string account_uuid, string password_hash, AccountCreateRequest model, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<bool> IsExistEmail(string email, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<string> VerifiedAccount(string email, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<LoginEntity?> GetAccount(string email, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<string> GetAccountUuid(string email, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task UpdatePassword(string account_uuid, string passwordHash, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
    }
}
