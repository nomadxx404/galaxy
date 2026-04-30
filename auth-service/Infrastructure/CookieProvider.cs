using auth_service.Config;
using auth_service.DTOs.Entity;

namespace auth_service.Infrastructure
{
    public interface ICookieProvider
    {
        void SetAuthCookies(string accessToken, string refreshToken);
        void ClearAuthCookies();
    }

    public class CookieProvider : ICookieProvider
    {
        private readonly IHttpContextAccessor _httpContextAccessor;
        private readonly AppConfig _appConfig;

        public CookieProvider(
            IHttpContextAccessor httpContextAccessor,
            AppConfig appConfig)
        {
            _httpContextAccessor = httpContextAccessor;
            _appConfig = appConfig;
        }

        public void ClearAuthCookies()
        {
            var context = _httpContextAccessor.HttpContext;
            if (context == null) return;

            var cookieOptions = new CookieOptions
            {
                HttpOnly = true,
                Secure = true,
                SameSite = SameSiteMode.Strict,
                Expires = DateTimeOffset.UtcNow.AddDays(-1)
            };

            context.Response.Cookies.Delete("access_token", cookieOptions);
            context.Response.Cookies.Delete("refresh_token", cookieOptions);
        }

        public void SetAuthCookies(string accessToken, string refreshToken)
        {
            var context = _httpContextAccessor.HttpContext;
            if (context == null) return;

            var accessOptionsAccess = new CookieOptions
            {
                HttpOnly = true,
                Secure = true,
                SameSite = SameSiteMode.Strict,
                Expires = DateTimeOffset.UtcNow.AddMinutes(_appConfig.JwtConfig.JWT_ACCESS_MINUTES)
            };

            var accessOptionsRefresh = new CookieOptions
            {
                HttpOnly = true,
                Secure = true,
                SameSite = SameSiteMode.Strict,
                Expires = DateTimeOffset.UtcNow.AddDays(_appConfig.JwtConfig.JWT_REFRESH_DAYS)
            };

            context.Response.Cookies.Append("access_token", accessToken, accessOptionsAccess);
            context.Response.Cookies.Append("refresh_token", refreshToken, accessOptionsRefresh);
        }
    }
}
