namespace auth_service.Config
{
    public class ArgonConfig
    {
        public int MemorySize => Env.GetInt("PASSWORD_HASHER_MEMORY_SIZE");
        public int Iterations => Env.GetInt("PASSWORD_HASHER_ITERATIONS");
        public int DegreeOfParallelism => Env.GetInt("PASSWORD_HASHER_DEGREE_OF_PARALLELISM");
        public int SaltLength => Env.GetInt("PASSWORD_HASHER_SALT_LENGTH");
        public int HashLength => Env.GetInt("PASSWORD_HASHER_HASH_LENGTH");
    }
}
