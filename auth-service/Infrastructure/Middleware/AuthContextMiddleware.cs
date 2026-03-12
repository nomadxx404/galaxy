namespace auth_service.Infrastructure.Middleware
{
    public class AuthContextMiddleware
    {
        private readonly RequestDelegate _next;
        private readonly string[] _publicPaths =
        {
            "/api/auth/login",
            "/api/auth/register",
            "/api/auth/forgot-password",
            "/api/auth/reset-password",
            "/api/auth/reset-otp-confirm",
            "/api/auth/otp-confirm",            
            "/scalar",
            "/openapi",
            "/swagger",
            "/favicon.ico"
        };

        public AuthContextMiddleware(RequestDelegate next)
        {
            _next = next;
        }

        public async Task InvokeAsync(HttpContext context)
        {
            //var internalToken = context.Request.Headers["X-Internal-Token"].ToString();

            //if (internalToken != "my-super-secret")
            //{
            //    context.Response.StatusCode = 403;
            //    context.Response.ContentType = "application/json";

            //    var failureResult = Result<string>.Failure(403, "Отказ в доступе");

            //    await context.Response.WriteAsJsonAsync(failureResult);
            //    return;
            //}

            var path = context.Request.Path.Value?.ToLower() ?? string.Empty;

            if (_publicPaths.Any(p => path.StartsWith(p.ToLower())))
            {
                await _next(context);
                return;
            }

            if (path.Contains("/refresh-token") || path.Contains("/logout"))
            {
                if (!context.Request.Cookies.ContainsKey("refresh_token"))
                {
                    context.Response.StatusCode = 401;
                    context.Response.ContentType = "application/json";

                    var failureResult = Result<string>.Failure(401, "Токен не найден");

                    await context.Response.WriteAsJsonAsync(failureResult);
                    return;
                }
            }

            var accountUuid = context.Request.Headers["X-Account-Uuid"].ToString();

            if (string.IsNullOrWhiteSpace(accountUuid))
            {
                context.Response.StatusCode = 401;
                context.Response.ContentType = "application/json";

                var failureResult = Result<string>.Failure(401, "Сессия не найдена. Пожалуйста, войдите снова");

                await context.Response.WriteAsJsonAsync(failureResult);

                return;
            }

            await _next(context);
        }
    }
}
