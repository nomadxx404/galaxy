namespace auth_service.Infrastructure
{
    public interface IUserContext
    {
        string? Account_uuid { get; }
        string? RefreshToken { get; }
    }

    public class UserContext : IUserContext
    {
        private readonly IHttpContextAccessor _httpContextAccessor;
        private const string HeaderName = "X-Account-Uuid";

        public UserContext(IHttpContextAccessor httpContextAccessor)
        {
            _httpContextAccessor = httpContextAccessor;
        }

        public string? Account_uuid =>
            _httpContextAccessor.HttpContext?.Request.Headers[HeaderName].ToString().Trim();

        public string? RefreshToken =>
        _httpContextAccessor.HttpContext?.Request.Cookies["refresh_token"];
    }
}
