using auth_service.Config;
using auth_service.Data;
using auth_service.Infrastructure;
using auth_service.Models.Request;
using auth_service.Repositories.Interfaces;
using auth_service.Services.Interfaces;

namespace auth_service.Services
{
    public class AuthService : IAuthService
    {
        private readonly PostgresConnection _conn;
        private readonly RedisConnection _connRedis;
        private readonly AppConfig _appConfig;
        private readonly IAuthRepository _authRepository;
        private readonly IPasswordHasher _passwordHasher;
        private readonly IJwtProvider _jwtProvider;
        private readonly ICookieProvider _cookieProvider;
        private readonly IUserContext _userContext;

        public AuthService(
            PostgresConnection conn,
            RedisConnection connRedis,
            AppConfig appConfig,
            IAuthRepository authRepository,
            IPasswordHasher passwordHasher,
            IJwtProvider jwtProvider,
            ICookieProvider cookieProvider,
            IUserContext userContext)
        {
            _conn = conn;
            _connRedis = connRedis;
            _appConfig = appConfig;
            _authRepository = authRepository;
            _passwordHasher = passwordHasher;
            _jwtProvider = jwtProvider;
            _cookieProvider = cookieProvider;
            _userContext = userContext;
        }

        public async Task<Result<EmailResponse>> Register(AccountCreateRequest accountCreateRequest)
        {
            var conn = await _conn.OpenConnectionAsync();

            bool isExistEmail = await _authRepository.IsExistEmail(email: accountCreateRequest.Email, conn);

            if (!isExistEmail)
                return Result<EmailResponse>.Failure(409, "Аккаунт с такой почтой уже существует");


            await _authRepository.AccountCreate(
                account_uuid: Guid.CreateVersion7().ToString(),
                password_hash: _passwordHasher.HashPassword(accountCreateRequest.Password),
                model: accountCreateRequest,
                conn: conn);

            await SaveAndSendOtp(accountCreateRequest.Email);

            //TODO: 
            //Kafka - create event send code email

            return Result<EmailResponse>.Success(201, "Аккаунт создан, код верификации отправлен на почту", new EmailResponse(accountCreateRequest.Email));
        }

        public async Task<Result<EmailResponse>> OtpConfirm(OtpConfirmRequest otpConfirmRequest)
        {
            string keyRedis = $"auth:otp:email_verify:{otpConfirmRequest.Email}";

            var codeRedis = await _connRedis.Database.StringGetAsync(keyRedis);

            if (!codeRedis.HasValue)
                return Result<EmailResponse>.Failure(410, "Срок действия кода истек. Пожалуйста, запросите новый код", new EmailResponse(otpConfirmRequest.Email));

            if (codeRedis != otpConfirmRequest.OtpCode)
                return Result<EmailResponse>.Failure(404, "Неверный код");

            await _connRedis.Database.KeyDeleteAsync(keyRedis);
            await _authRepository.VerifiedAccount(
                email: otpConfirmRequest.Email, 
                conn: await _conn.OpenConnectionAsync());

            //TODO: 
            //Kafka - create event for loggin

            return Result<EmailResponse>.Success(200, "Аккаунт успешно верифицирован", new EmailResponse(otpConfirmRequest.Email));
        }

        public async Task<Result<EmailResponse>> ResetOtpConfirm(EmailRequest emailRequest)
        {
            await SaveAndSendOtp(emailRequest.Email);

            //TODO: 
            //Kafka - create event for loggin

            return Result<EmailResponse>.Success(200, "Код отправлен", new EmailResponse(emailRequest.Email));
        }

        public async Task<Result<string>> Login(LoginRequest loginRequest)
        {
            var account = await _authRepository.GetAccount(
                email: loginRequest.Email, 
                conn: await _conn.OpenConnectionAsync());

            if (account == null || !_passwordHasher.VerifyPassword(loginRequest.Password, account.Password_hash))
                return Result<string>.Failure(403, "Неверный логин или пароль");

            if (!account.Is_verified)
                return Result<string>.Failure(401, "Аккаунт не верифицирован. Пожалуйста, подтвердите ваш адрес электронной почты, чтобы продолжить");

            var tokenResult = await IssueTokens(account.Account_uuid);

            if (!tokenResult.IsSuccess)
                return tokenResult;

            //TODO: 
            //Kafka - create event for loggin

            return Result<string>.Success(200, "Выполнен вход");
        }

        public async Task<Result<string>> Logout()
        {
            string refreshKey = $"auth:refresh:{_userContext.RefreshToken}";
            string userSessionsKey = $"auth:sessions:{_userContext.Account_uuid}";

            var batch = _connRedis.Database.CreateBatch();

            var t1 = batch.KeyDeleteAsync(refreshKey);
            var t2 = batch.SetRemoveAsync(userSessionsKey, _userContext.RefreshToken);

            batch.Execute();
            await Task.WhenAll(t1, t2);

            _cookieProvider.ClearAuthCookies();

            //TODO: 
            //Kafka - create event for loggin

            return Result<string>.Success(200, "Выполнен выход");
        }

        public async Task<Result<string>> FullLogout()
        {
            string userSessionsKey = $"auth:sessions:{_userContext.Account_uuid}";

            var allTokens = await _connRedis.Database.SetMembersAsync(userSessionsKey);

            if (allTokens.Length > 0)
            {
                var batch = _connRedis.Database.CreateBatch();

                foreach (var token in allTokens)
                {
                    batch.KeyDeleteAsync($"auth:refresh:{token}");
                }

                batch.KeyDeleteAsync(userSessionsKey);

                batch.Execute();
            }

            _cookieProvider.ClearAuthCookies();

            return Result<string>.Success(200, "Выполнен выход на всех устройствах");
        }

        public async Task<Result<string>> RefreshToken()
        {
            string oldRefreshKey = $"auth:refresh:{_userContext.RefreshToken}";

            var account_uuid = await _connRedis.Database.StringGetAsync(oldRefreshKey);

            if (account_uuid.IsNull)
                return Result<string>.Failure(401, "Сессия истекла. Пожалуйста, войдите снова.");

            string userSessionsKey = $"auth:sessions:{account_uuid}";

            var batch = _connRedis.Database.CreateBatch();
            var t1 = batch.KeyDeleteAsync(oldRefreshKey);
            var t2 = batch.SetRemoveAsync(userSessionsKey, _userContext.RefreshToken);
            batch.Execute();
            await Task.WhenAll(t1, t2);

            var tokenResult = await IssueTokens(account_uuid.ToString());

            if (!tokenResult.IsSuccess)
                return tokenResult;

            //TODO: 
            //Kafka - create event for loggin

            return Result<string>.Success(200, "");
        }


        public async Task<Result<string>> ForgotPassword(EmailRequest emailRequest)
        {            
            var account = await _authRepository.GetAccount(
                email: emailRequest.Email, 
                conn: await _conn.OpenConnectionAsync());

            if (account == null)
                return Result<string>.Failure(404, "Аккаунт с такой почтой не найден");

            string resetToken = Guid.NewGuid().ToString("N");
            string resetKey = $"auth:reset:{resetToken}";

            await _connRedis.Database.StringSetAsync(resetKey, account.Account_uuid, TimeSpan.FromMinutes(_appConfig.AuthConfig.RESET_PASSWORD_TOKEN_ACCESS_MINUTES));

            //TODO: 
            //Kafka - create event send token email

            return Result<string>.Success(200, "Ссылка для восстановления пароля отправдена на почту");
        }


        public async Task<Result<string>> ResetPassword(ResetPasswordRequest resetPasswordRequest)
        {
            string resetKey = $"auth:reset:{resetPasswordRequest.Token}";

            var account_uuid = await _connRedis.Database.StringGetAsync(resetKey);

            if (account_uuid.IsNull)
                return Result<string>.Failure(410, "Срок действия ссылки истек или она недействительна");

            await _authRepository.UpdatePassword(
                account_uuid: account_uuid.ToString(), 
                passwordHash: _passwordHasher.HashPassword(resetPasswordRequest.NewPassword),
                conn: await _conn.OpenConnectionAsync());

            string userSessionsKey = $"auth:sessions:{account_uuid}";

            var sessions = await _connRedis.Database.SetMembersAsync(userSessionsKey);

            var batch = _connRedis.Database.CreateBatch();

            foreach (var session in sessions)
            {
                batch.KeyDeleteAsync($"auth:refresh:{session}");
            }
            batch.KeyDeleteAsync(userSessionsKey);
            batch.KeyDeleteAsync(resetKey);
            batch.Execute();

            //TODO: 
            //Kafka - create event for loggin

            return Result<string>.Success(200, "Пароль успешно изменен. Войдите с новым паролем");
        }

        private async Task SaveAndSendOtp(string email)
        {
            await _connRedis.Database.StringSetAsync(
                $"auth:otp:email_verify:{email}",
                CodeGenerator.GenerateSixDigitCode(),
                TimeSpan.FromMinutes(_appConfig.AuthConfig.OTP_CODE_ACCESS_MINUTES));
        }

        private async Task<Result<string>> IssueTokens(string account_uuid)
        {
            try
            {
                string accessToken = _jwtProvider.GenerateAccessToken(account_uuid);
                string refreshToken = _jwtProvider.GenerateRefreshToken();

                string refreshKey = $"auth:refresh:{refreshToken}";
                string userSessionsKey = $"auth:sessions:{account_uuid}";

                var db = _connRedis.Database;
                var batch = db.CreateBatch();

                var t1 = batch.StringSetAsync(refreshKey, account_uuid, TimeSpan.FromDays(_appConfig.JwtConfig.JWT_REFRESH_DAYS));
                var t2 = batch.SetAddAsync(userSessionsKey, refreshToken);
                var t3 = batch.KeyExpireAsync(userSessionsKey, TimeSpan.FromDays(_appConfig.JwtConfig.JWT_REFRESH_DAYS));

                batch.Execute();
                await Task.WhenAll(t1, t2, t3);

                _cookieProvider.SetAuthCookies(accessToken: accessToken, refreshToken: refreshToken);

                return Result<string>.Success(200, "");
            }
            catch (Exception ex)
            {
                return Result<string>.Failure(500, "Ошибка сервера при создании сессии");
            }
        }
    }
}
