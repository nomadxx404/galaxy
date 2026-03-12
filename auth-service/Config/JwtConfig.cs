namespace auth_service.Config
{
    public class JwtConfig
    {
        public string JWT_KEY => Env.Get("JWT_KEY");
        public string JWT_ISSUER => Env.Get("JWT_ISSUER");
        public string JWT_AUDIENCE => Env.Get("JWT_AUDIENCE");
        public int JWT_ACCESS_MINUTES => Env.GetInt("JWT_ACCESS_MINUTES");
        public int JWT_REFRESH_DAYS => Env.GetInt("JWT_REFRESH_DAYS");
    }
}
