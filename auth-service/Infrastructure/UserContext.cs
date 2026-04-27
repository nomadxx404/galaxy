namespace auth_service.Infrastructure
{
    public interface IUserContext
    {
        string? Account_uuid { get; }
        string? RefreshToken { get; }
        string RequestId { get; }
        string Method { get; }
        string Path { get; }
    }

    public class UserContext : IUserContext
    {
        private readonly IHttpContextAccessor _httpContextAccessor;
        private const string AccountHeader = "X-Account-Uuid";
        private const string RequestIdHeader = "X-Request-Id";
        private const string RefreshTokenCookies = "refresh_token";

        public UserContext(IHttpContextAccessor httpContextAccessor)
        {
            _httpContextAccessor = httpContextAccessor;
        }

        public string? Account_uuid =>
            _httpContextAccessor.HttpContext?.Request.Headers[AccountHeader].ToString().Trim();

        public string? RefreshToken =>
            _httpContextAccessor.HttpContext?.Request.Cookies[RefreshTokenCookies];

        public string RequestId =>
            !string.IsNullOrWhiteSpace(_httpContextAccessor.HttpContext?.Request.Headers[RequestIdHeader])
            ? _httpContextAccessor.HttpContext!.Request.Headers[RequestIdHeader].ToString()
            : (_httpContextAccessor.HttpContext?.TraceIdentifier ?? Guid.NewGuid().ToString());

        public string Method =>
            _httpContextAccessor.HttpContext?.Request.Method ?? string.Empty;

        public string Path =>
            _httpContextAccessor.HttpContext?.Request.Path.ToString() ?? string.Empty;
    }
}
