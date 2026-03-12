using System.ComponentModel.DataAnnotations;

namespace auth_service.Models.Request
{
    public record AccountCreateRequest(
        [Required] string Name,
        [Required] string Nickname,
        [Required, EmailAddress] string Email,
        [Required] string Password,
        [Required] int AvatarFileId);

    public record OtpConfirmRequest(
        [Required, EmailAddress] string Email,
        [Required] string OtpCode);

    public record EmailRequest(
        [Required, EmailAddress] string Email);

    public record LoginRequest(
        [Required, EmailAddress] string Email,
        [Required] string Password);

    public record ResetPasswordRequest(
        [Required] string Token,
        [Required] string NewPassword);
}
