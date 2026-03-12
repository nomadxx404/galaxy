namespace auth_service.Infrastructure.Middleware
{
    public class GlobalExceptionMiddleware
    {
        private readonly RequestDelegate _next;
        private readonly IHostEnvironment _environment;

        public GlobalExceptionMiddleware(
            RequestDelegate next,
            IHostEnvironment environment)
        {
            _next = next;
            _environment = environment;
        }

        public async Task InvokeAsync(HttpContext context)
        {
            try
            {
                await _next(context);
            }
            catch (Exception ex)
            {
                context.Response.ContentType = "application/json";
                context.Response.StatusCode = StatusCodes.Status500InternalServerError;

                var response = new
                {
                    success = false,
                    statusCode = context.Response.StatusCode,
                    message = "Внутренняя ошибка сервера. Мы уже работаем над этим.",
                    detail = _environment.IsDevelopment() ? ex.Message : null
                };

                await context.Response.WriteAsJsonAsync(response);
            }
        }
    }
}
