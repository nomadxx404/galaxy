namespace auth_service.Config
{
    public class AuthConfig
    {
        public int OTP_CODE_ACCESS_MINUTES => Env.GetInt("OTP_CODE_ACCESS_MINUTES");
        public int RESET_PASSWORD_TOKEN_ACCESS_MINUTES => Env.GetInt("RESET_PASSWORD_TOKEN_ACCESS_MINUTES");
    }
}
