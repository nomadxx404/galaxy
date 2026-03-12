namespace auth_service.DTO.Response
{
    public record AccountResponse(
        string Name,
        string Nickname,
        string Email,
        int Avatar_file_id);

    public record AccountUpdateResponse(
       string Name,
       string Nickname,
       int Avatar_file_id);

    public record AccountBatchResponse(
       string Account_uuid,
       string Name,
       string Nickname,
       int Avatar_file_id);
}
