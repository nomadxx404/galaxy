namespace auth_service.Infrastructure.Kafka
{
    public class Events
    {
        public const string Register = "Register";
        public const string OtpConfirm = "OtpConfirm";
        public const string ResetOtpConfirm = "ResetOtpConfirm";
        public const string Login = "Login";
        public const string Logout = "Logout";
        public const string FullLogout = "FullLogout";
        public const string ForgotPassword = "ForgotPassword";
        public const string ResetPassword = "ResetPassword";

        public const string AccountUpdated = "AccountUpdated";
        public const string ChangePassword = "ChangePassword";
        public const string AccountDeleted = "AccountDeleted";
    }
}
