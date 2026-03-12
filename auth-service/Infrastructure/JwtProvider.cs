using auth_service.Config;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace auth_service.Infrastructure
{
    public interface IJwtProvider
    {
        string GenerateAccessToken(string accountUuid);
        string GenerateRefreshToken();
    }

    public class JwtProvider(AppConfig config) : IJwtProvider
    {
        public string GenerateAccessToken(string accountUuid)
        {
            var claims = new[]
            {
                new Claim(JwtRegisteredClaimNames.Sub, accountUuid),
                new Claim(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
            };
            var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(config.JwtConfig.JWT_KEY));
            var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

            var token = new JwtSecurityToken(
                issuer: config.JwtConfig.JWT_ISSUER,
                audience: config.JwtConfig.JWT_AUDIENCE,
                claims: claims,
                expires: DateTime.UtcNow.AddMinutes(config.JwtConfig.JWT_ACCESS_MINUTES),
                signingCredentials: creds
            );

            return new JwtSecurityTokenHandler().WriteToken(token);
        }

        public string GenerateRefreshToken() => Guid.NewGuid().ToString("N");
    }
}
