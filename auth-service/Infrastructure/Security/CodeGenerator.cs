using System.Security.Cryptography;

namespace auth_service.Infrastructure
{
    public static class CodeGenerator
    {
        public static string GenerateSixDigitCode()
        {
            int code = RandomNumberGenerator.GetInt32(100000, 1000000);
            return code.ToString();
        }
    }
}
