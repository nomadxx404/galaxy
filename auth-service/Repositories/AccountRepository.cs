using auth_service.DTO.Entity;
using auth_service.DTO.Request;
using auth_service.DTO.Response;
using auth_service.Repositories.Interfaces;
using Dapper;
using Npgsql;
using System.Data;

namespace auth_service.Repositories
{
    public class AccountRepository : IAccountRepository
    {
        public async Task<AccountResponse?> GetAccount(string account_uuid, NpgsqlConnection conn)
        {
            return await conn.QueryFirstOrDefaultAsync<AccountResponse?>(
                """
                SELECT
                    name,
                    nickname,
                    email,
                    avatar_file_id
                FROM auth.account
                WHERE account_uuid = @account_uuid
                """,
                new
                {
                    account_uuid = account_uuid,
                });
        }

        public async Task<AccountUpdateResponse?> UpdateProfile(string account_uuid, AccountUpdateRequest request, NpgsqlConnection conn)
        {
            return await conn.QuerySingleOrDefaultAsync<AccountUpdateResponse?>(
                """
                UPDATE auth.account
                SET
                    name = COALESCE(@name, name),
                    nickname = COALESCE(@nickname, nickname),
                    avatar_file_id = COALESCE(@avatar_file_id, avatar_file_id),
                    updated_at = NOW()
                WHERE account_uuid = @account_uuid
                RETURNING 
                    name, nickname, avatar_file_id;
                """,
                new
                {
                    account_uuid = account_uuid,
                    name = request.Name,
                    nickname = request.Nickname,
                    avatar_file_id = request.AvatarFileId
                });
        }

        public async Task<string> GetCurrentPassword(string account_uuid, NpgsqlConnection conn)
        {
            return await conn.QueryFirstOrDefaultAsync<string>(
                """
                SELECT password_hash
                FROM auth.account
                WHERE account_uuid = @account_uuid
                """,
                new
                {
                    account_uuid = account_uuid,
                });
        }

        public async Task ChangePassword(string account_uuid, string passwordHash, NpgsqlConnection conn)
        {
            await conn.ExecuteAsync(
                """
                UPDATE auth.account
                SET password_hash = @password_hash
                WHERE account_uuid = @account_uuid
                """,
                new
                {
                    account_uuid = account_uuid,
                    password_hash = passwordHash
                });
        }

        public async Task DeleteAccount(string account_uuid, NpgsqlConnection conn)
        {
            await conn.ExecuteAsync(
                """
                UPDATE auth.account
                SET is_active = false
                WHERE account_uuid  = @account_uuid
                """,
                new
                {
                    account_uuid = account_uuid,
                });
        }

        public async Task<IEnumerable<AccountBatchResponse>> GetAccountsByUuids(IEnumerable<string> account_uuids, NpgsqlConnection conn)
        {
            return await conn.QueryAsync<AccountBatchResponse>(
                """
                SELECT 
                    account_uuid, 
                    name, 
                    nickname, 
                    avatar_file_id
                FROM auth.account
                WHERE 
                    account_uuid = ANY(@account_uuids)
                """, 
                new 
                { 
                    account_uuids = account_uuids.ToArray() 
                });
        }
    }
}
