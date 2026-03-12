using auth_service.DTO.Request;
using auth_service.Infrastructure;
using auth_service.Services.Interfaces;
using Microsoft.AspNetCore.Mvc;

namespace auth_service.Controllers
{
    [Route("api/profile")]
    [ApiController]
    public class AccountController : ControllerBase
    {
        private readonly IAccountService _accountService;
        public AccountController(IAccountService accountService) => _accountService = accountService;

        [HttpGet]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status404NotFound)]
        public async Task<IActionResult> GetProfile()
        {
            var result = await _accountService.GetProfile();
            return result.ToActionResult(this);
        }


        [HttpPatch]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status404NotFound)]
        public async Task<IActionResult> UpdateProfile(AccountUpdateRequest request)
        {
            var result = await _accountService.UpdateProfile(request);
            return result.ToActionResult(this);
        }


        [HttpPost("change-password")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status404NotFound)]
        public async Task<IActionResult> ChangePassword(AccountChangePasswordRequest request)
        {
            var result = await _accountService.ChangePassword(request);
            return result.ToActionResult(this);
        }


        [HttpDelete]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status404NotFound)]
        public async Task<IActionResult> DeleteAccount()
        {
            var result = await _accountService.DeleteAccount();
            return result.ToActionResult(this);
        }


        [HttpPost("batch")]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status200OK)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status400BadRequest)]
        [ProducesResponseType(typeof(ApiMessageResponse), StatusCodes.Status401Unauthorized)]
        public async Task<IActionResult> GetProfilesBatch(AccountBatchRequest request)
        {
            var result = await _accountService.GetProfilesBatch(request);
            return StatusCode(result.StatusCode, result);
        }
    }
}
