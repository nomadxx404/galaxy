using auth_service.Config;
using auth_service.Data;
using auth_service.Repositories;
using auth_service.Repositories.Interfaces;
using auth_service.Services;
using auth_service.Services.Interfaces;

namespace auth_service.Infrastructure
{
    public static class AddServiceExtensions
    {
        public static IServiceCollection AddInfrastructure(this IServiceCollection services)
        {
            AppConfig config = new AppConfig();

            services.AddSingleton(config);
            services.AddSingleton(config.PostgresConfig);
            services.AddSingleton(config.RedisConfig);
            services.AddSingleton(config.JwtConfig);
            services.AddSingleton(config.ArgonConfig);
            services.AddSingleton(config.AuthConfig);


            services.AddSingleton<RedisConnection>();
            services.AddSingleton<PostgresConnection>();


            services.AddScoped<IPasswordHasher, Argon2PasswordHasher>();
            services.AddScoped<IJwtProvider, JwtProvider>();
            services.AddScoped<ICookieProvider, CookieProvider>();
            services.AddScoped<IUserContext, UserContext>();


            services.AddScoped<IAuthRepository, AuthRepository>();
            services.AddScoped<IAuthService, AuthService>();


            services.AddScoped<IAccountRepository, AccountRepository>();
            services.AddScoped<IAccountService, AccountService>();


            return services;
        }
    }
}
