using auth_service.Infrastructure;
using auth_service.Models.Request;
using auth_service.Services.Interfaces;
using Microsoft.AspNetCore.Mvc;

namespace auth_service.Controllers
{
    [Route("api/auth")]
    [ApiController]
    public class AuthController : Controller
    {
        private readonly IAuthService _authService;
        public AuthController(IAuthService authService) => _authService = authService;

        [HttpPost("register")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status201Created)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status409Conflict)]
        public async Task<IActionResult> Register(AccountCreateRequest request)
        {
            var result = await _authService.Register(request);
            return result.ToActionResult(this);
        }


        [HttpPost("otp-confirm")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status404NotFound)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status410Gone)]
        public async Task<IActionResult> OtpConfirm(OtpConfirmRequest request)
        {
            var result = await _authService.OtpConfirm(request);
            return result.ToActionResult(this);
        }


        [HttpPost("reset-otp-confirm")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        public async Task<IActionResult> ResetOtpConfirm(EmailRequest request)
        {
            var result = await _authService.ResetOtpConfirm(request);
            return result.ToActionResult(this);
        }


        [HttpPost("login")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status403Forbidden)]
        public async Task<IActionResult> Login(LoginRequest request)
        {
            var result = await _authService.Login(request);
            return result.ToActionResult(this);
        }


        [HttpPost("logout")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        public async Task<IActionResult> Logout()
        {
            var result = await _authService.Logout();
            return result.ToActionResult(this);
        }


        [HttpPost("full-logout")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        public async Task<IActionResult> FullLogout()
        {
            var result = await _authService.FullLogout();
            return result.ToActionResult(this);
        }


        [HttpPost("refresh-token")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        public async Task<IActionResult> RefreshToken()
        {
            var result = await _authService.RefreshToken();
            return result.ToActionResult(this);
        }

        [HttpPost("forgot-password")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        public async Task<IActionResult> ForgotPassword(EmailRequest request)
        {
            var result = await _authService.ForgotPassword(request);
            return result.ToActionResult(this);
        }

        [HttpPost("reset-password")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        public async Task<IActionResult> ResetPassword(ResetPasswordRequest request)
        {
            var result = await _authService.ResetPassword(request);
            return result.ToActionResult(this);
        }
    }
}
