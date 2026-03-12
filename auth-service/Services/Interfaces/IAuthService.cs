using auth_service.Infrastructure;
using auth_service.Models.Request;

namespace auth_service.Services.Interfaces
{
    public interface IAuthService
    {
        public Task<Result<EmailResponse>> Register(AccountCreateRequest AccountCreateRequest);
        public Task<Result<EmailResponse>> OtpConfirm(OtpConfirmRequest OtpConfirmRequest);
        public Task<Result<EmailResponse>> ResetOtpConfirm(EmailRequest emailRequest);
        public Task<Result<string>> Login(LoginRequest loginRequest);
        public Task<Result<string>> Logout();
        public Task<Result<string>> FullLogout();
        public Task<Result<string>> RefreshToken();
        public Task<Result<string>> ForgotPassword(EmailRequest emailRequest);
        public Task<Result<string>> ResetPassword(ResetPasswordRequest resetPasswordRequest);
    }
}
