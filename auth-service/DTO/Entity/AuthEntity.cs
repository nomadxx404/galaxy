using System.ComponentModel.DataAnnotations;

namespace auth_service.DTOs.Entity
{
    public record LoginEntity(
        string Account_uuid,
        string Password_hash,
        bool Is_verified);

    
}
