using auth_service.DTO.Request;
using auth_service.DTO.Response;
using auth_service.Infrastructure;

namespace auth_service.Services.Interfaces
{
    public interface IAccountService
    {
        public Task<Result<AccountResponse>> GetProfile();
        public Task<Result<AccountUpdateResponse>> UpdateProfile(AccountUpdateRequest accountRequest);
        public Task<Result<string>> ChangePassword(AccountChangePasswordRequest accountChangePasswordRequest);
        public Task<Result<string>> DeleteAccount();
        public Task<Result<IEnumerable<AccountBatchResponse>>> GetProfilesBatch(AccountBatchRequest accountBatchRequest);
    }
}
