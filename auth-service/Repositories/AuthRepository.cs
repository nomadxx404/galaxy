using auth_service.DTOs.Entity;
using auth_service.Models.Request;
using auth_service.Repositories.Interfaces;
using Dapper;
using Npgsql;

namespace auth_service.Repositories
{
    public class AuthRepository : IAuthRepository
    {
        public async Task<bool> IsExistEmail(string email, NpgsqlConnection conn)
        {
            return await conn.QueryFirstOrDefaultAsync<bool>(
               """
               SELECT COUNT(*) = 0
               FROM auth.account
               WHERE 
                email = @email
               """,
               new
               {
                   email = email
               });
        }

        public async Task AccountCreate(string account_uuid, string passwordHash, AccountCreateRequest model, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                INSERT INTO auth.account
                (
                    account_uuid,
                    name,
                    nickname,
                    email,
                    password_hash,
                    avatar_file_id
                )
                VALUES
                (
                    @account_uuid,
                    @name,
                    @nickname,
                    @email,
                    @password_hash,
                    @avatar_file_id
                );
                """,
                new
                {
                    account_uuid = account_uuid,
                    name = model.Name,
                    nickname = model.Nickname,
                    email = model.Email,
                    password_hash = passwordHash,
                    avatar_file_id = model.AvatarFileId
                });
        }

        public async Task CreateOtpCode(string email, string otp_code, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                INSERT INTO auth.otp_confirm
                (
                    email,
                    otp_code
                )
                VALUES
                (
                    @email,
                    @otp_code
                )
                """,
                new
                {
                    email = email,
                    otp_code = otp_code
                },
                transaction);
        }

        public async Task VerifiedAccount(string email, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                UPDATE auth.account
                SET is_verified = true
                WHERE email = @email
                """,
                new
                {
                    email = email,
                });
        }


        public async Task<LoginEntity?> GetAccount(string email, NpgsqlConnection conn)
        {
            return await conn.QueryFirstOrDefaultAsync<LoginEntity?>(
                """
                SELECT
                    account_uuid,
                    password_hash,
                    is_verified
                FROM auth.account
                WHERE 
                    email = @email 
                """,
                new
                {
                    email = email
                });            
        }

        public async Task UpdatePassword(string account_uuid ,string passwordHash, NpgsqlConnection conn)
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
    }
}
