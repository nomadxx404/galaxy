using auth_service.Data;
using auth_service.DTO.Request;
using auth_service.DTO.Response;
using auth_service.Infrastructure;
using auth_service.Infrastructure.Kafka;
using auth_service.Repositories.Interfaces;
using auth_service.Services.Interfaces;

namespace auth_service.Services
{
    public class AccountService : IAccountService
    {
        private readonly PostgresConnection _conn;
        private readonly IAccountRepository _accountRepository;
        private readonly IUserContext _userContext;
        private readonly IPasswordHasher _passwordHasher;
        private readonly IAuthService _authService;
        private readonly IOutboxEmitter _outboxEmitter;

        public AccountService(
            PostgresConnection conn,
            IAccountRepository accountRepository,
            IUserContext userContext,
            IPasswordHasher passwordHasher,
            IAuthService authService,
            IOutboxEmitter outboxEmitter)
        {
            _conn = conn;
            _accountRepository = accountRepository;
            _userContext = userContext;
            _passwordHasher = passwordHasher;
            _authService = authService;
            _outboxEmitter = outboxEmitter;
        }

        public async Task<Result<AccountResponse>> GetProfile()
        {
            var account = await _accountRepository.GetAccount(
                account_uuid: _userContext.Account_uuid,
                conn: await _conn.OpenConnectionAsync());

            if (account == null)
                return Result<AccountResponse>.Failure(404, "Профиль не найден");

            return Result<AccountResponse>.Success(200, "Успешно", account);
        }

        public async Task<Result<AccountUpdateResponse>> UpdateProfile(AccountUpdateRequest accountUpdateRequest)
        {
            using var conn = await _conn.OpenConnectionAsync();
            using var transaction = await conn.BeginTransactionAsync();

            var account = await _accountRepository.UpdateProfile(
                account_uuid: _userContext.Account_uuid,
                request: accountUpdateRequest,
                conn: conn,
                transaction: transaction);

            if (account == null)
                return Result<AccountUpdateResponse>.Failure(404, "Ошибка обновления профиля");

            var payload = new
            {
                account_uuid = _userContext.Account_uuid,
            };

            await _outboxEmitter.EmitOutbox(
                eventType: Events.AccountUpdated,
                payload: payload,
                conn: conn,
                transaction: transaction);

            await transaction.CommitAsync();

            return Result<AccountUpdateResponse>.Success(200, "Профиль обновлен успешно", account);
        }

        public async Task<Result<string>> ChangePassword(AccountChangePasswordRequest request)
        {
            if (request.CurrentPassword == request.NewPassword)
                return Result<string>.Failure(400, "Новый пароль не может совпадать со старым");

            using var conn = await _conn.OpenConnectionAsync();
            using var transaction = await conn.BeginTransactionAsync();

            string currentPasswordHash = await _accountRepository.GetCurrentPassword(
                account_uuid: _userContext.Account_uuid!,
                conn: conn,
                transaction: transaction);

            if (currentPasswordHash == null)
                return Result<string>.Failure(404, "Профиль не найден");

            bool verifyPassword = _passwordHasher.VerifyPassword(
                password: request.CurrentPassword,
                hash: currentPasswordHash);

            if (!verifyPassword)
                return Result<string>.Failure(403, "Текущий пароль введен неверно");

            string newHash = _passwordHasher.HashPassword(request.NewPassword);
            await _accountRepository.ChangePassword(
                account_uuid: _userContext.Account_uuid!,
                passwordHash: newHash,
                conn: conn,
                transaction:transaction);


            var payload = new
            {
                account_uuid = _userContext.Account_uuid,
            };

            await _outboxEmitter.EmitOutbox(
                eventType: Events.ChangePassword,
                payload: payload,
                conn: conn,
                transaction: transaction);

            await transaction.CommitAsync();

            await _authService.FullLogout();

            return Result<string>.Success(200, "Пароль обновлен успешно. Пожалуйста, войдите снова.");
        }

        public async Task<Result<string>> DeleteAccount()
        {
            using var conn = await _conn.OpenConnectionAsync();
            using var transaction = await conn.BeginTransactionAsync();

            await _accountRepository.DeleteAccount(
                account_uuid: _userContext.Account_uuid, 
                conn: conn,
                transaction: transaction);

            var payload = new
            {
                account_uuid = _userContext.Account_uuid,
            };

            await _outboxEmitter.EmitOutbox(
                eventType: Events.AccountDeleted,
                payload: payload,
                conn: conn,
                transaction: transaction);

            await transaction.CommitAsync();

            await _authService.FullLogout();

            return Result<string>.Success(200, "Аккант успешно удален");
        }

        public async Task<Result<IEnumerable<AccountBatchResponse>>> GetProfilesBatch(AccountBatchRequest accountBatchRequest)
        {
            if (accountBatchRequest.Account_uuids == null || !accountBatchRequest.Account_uuids.Any())
                return Result<IEnumerable<AccountBatchResponse>>.Success(200, "Список пуст");

            var profiles = await _accountRepository.GetAccountsByUuids(
                account_uuids: accountBatchRequest.Account_uuids,
                conn: await _conn.OpenConnectionAsync());

            return Result<IEnumerable<AccountBatchResponse>>.Success(200, "OK", profiles);
        }
    }
}
