namespace auth_service.Config
{
    public static class Env
    {
        public static string Get(string key)
        {
            var value = Environment.GetEnvironmentVariable(key);

            if (string.IsNullOrEmpty(value))
                throw new Exception($"Environment variable '{key}' not found");

            return value;
        }

        public static int GetInt(string key)
        {
            return int.Parse(Get(key));
        }
    }
}
