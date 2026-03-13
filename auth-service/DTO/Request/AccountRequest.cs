using System.ComponentModel.DataAnnotations;

namespace auth_service.DTO.Request
{
    public record AccountUpdateRequest(
        string? Name,
        string? Nickname,
        int? AvatarFileId);

    public record AccountChangePasswordRequest(
        [Required] string CurrentPassword,
        [Required] string NewPassword);

    public record AccountBatchRequest(
        List<string> Account_uuids
    );
}
