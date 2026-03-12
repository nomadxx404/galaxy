using auth_service.Config;
using Konscious.Security.Cryptography;
using System.Globalization;
using System.Security.Cryptography;
using System.Text;

namespace auth_service.Infrastructure
{
    public interface IPasswordHasher
    {
        string HashPassword(string password);
        bool VerifyPassword(string password, string hash);
    }

    public sealed class Argon2PasswordHasher : IPasswordHasher
    {
        private readonly ArgonConfig _options;

        public Argon2PasswordHasher(AppConfig options)
        {
            _options = options.ArgonConfig;
            ValidateOptions(_options);
        }

        public string HashPassword(string password)
        {
            if (string.IsNullOrEmpty(password))
                throw new ArgumentException("Password cannot be null or empty.", nameof(password));

            byte[] salt = RandomNumberGenerator.GetBytes(_options.SaltLength);
            byte[] passwordBytes = Encoding.UTF8.GetBytes(password);

            try
            {
                using var argon2 = new Argon2id(passwordBytes)
                {
                    Salt = salt,
                    DegreeOfParallelism = _options.DegreeOfParallelism,
                    MemorySize = _options.MemorySize,
                    Iterations = _options.Iterations
                };

                byte[] hash = argon2.GetBytes(_options.HashLength);

                return string.Create(CultureInfo.InvariantCulture,
                    $"$argon2id$v=19$m={_options.MemorySize},t={_options.Iterations},p={_options.DegreeOfParallelism}${Convert.ToBase64String(salt)}${Convert.ToBase64String(hash)}");
            }
            finally
            {
                CryptographicOperations.ZeroMemory(passwordBytes);
            }
        }

        public bool VerifyPassword(string password, string encodedHash)
        {
            if (string.IsNullOrEmpty(password) || string.IsNullOrWhiteSpace(encodedHash))
                return false;

            byte[] passwordBytes = Encoding.UTF8.GetBytes(password);

            try
            {
                if (!TryParseHash(encodedHash, out var parsed))
                    return false;

                using var argon2 = new Argon2id(passwordBytes)
                {
                    Salt = parsed.Salt,
                    DegreeOfParallelism = parsed.DegreeOfParallelism,
                    MemorySize = parsed.MemorySize,
                    Iterations = parsed.Iterations
                };

                byte[] computedHash = argon2.GetBytes(parsed.Hash.Length);
                return CryptographicOperations.FixedTimeEquals(computedHash, parsed.Hash);
            }
            catch
            {
                return false;
            }
            finally
            {
                CryptographicOperations.ZeroMemory(passwordBytes);
            }
        }

        private static void ValidateOptions(ArgonConfig options)
        {
            if (options.MemorySize < 19 * 1024)
                throw new ArgumentOutOfRangeException(nameof(options.MemorySize), "MemorySize must be at least 19456 KiB (19 MiB).");

            if (options.Iterations < 2)
                throw new ArgumentOutOfRangeException(nameof(options.Iterations), "Iterations must be at least 2.");

            if (options.DegreeOfParallelism < 1)
                throw new ArgumentOutOfRangeException(nameof(options.DegreeOfParallelism), "DegreeOfParallelism must be at least 1.");

            if (options.SaltLength < 16)
                throw new ArgumentOutOfRangeException(nameof(options.SaltLength), "SaltLength must be at least 16 bytes.");

            if (options.HashLength < 16)
                throw new ArgumentOutOfRangeException(nameof(options.HashLength), "HashLength must be at least 16 bytes.");
        }

        private static bool TryParseHash(string encodedHash, out ParsedHash parsed)
        {
            parsed = default!;

            var parts = encodedHash.Split('$', StringSplitOptions.None);
            if (parts.Length != 6)
                return false;

            if (!string.Equals(parts[1], "argon2id", StringComparison.Ordinal))
                return false;

            if (!string.Equals(parts[2], "v=19", StringComparison.Ordinal))
                return false;

            if (!TryParseParameters(parts[3], out int memorySize, out int iterations, out int degreeOfParallelism))
                return false;

            byte[] salt;
            byte[] hash;

            try
            {
                salt = Convert.FromBase64String(parts[4]);
                hash = Convert.FromBase64String(parts[5]);
            }
            catch (FormatException)
            {
                return false;
            }

            if (salt.Length < 16 || hash.Length < 16)
                return false;

            parsed = new ParsedHash
            {
                MemorySize = memorySize,
                Iterations = iterations,
                DegreeOfParallelism = degreeOfParallelism,
                Salt = salt,
                Hash = hash
            };

            return true;
        }

        private static bool TryParseParameters(string parameters, out int memorySize, out int iterations, out int degreeOfParallelism)
        {
            memorySize = 0;
            iterations = 0;
            degreeOfParallelism = 0;

            var chunks = parameters.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
            if (chunks.Length != 3)
                return false;

            foreach (var chunk in chunks)
            {
                if (chunk.StartsWith("m=", StringComparison.Ordinal))
                {
                    if (!int.TryParse(chunk.AsSpan(2), NumberStyles.None, CultureInfo.InvariantCulture, out memorySize))
                        return false;
                }
                else if (chunk.StartsWith("t=", StringComparison.Ordinal))
                {
                    if (!int.TryParse(chunk.AsSpan(2), NumberStyles.None, CultureInfo.InvariantCulture, out iterations))
                        return false;
                }
                else if (chunk.StartsWith("p=", StringComparison.Ordinal))
                {
                    if (!int.TryParse(chunk.AsSpan(2), NumberStyles.None, CultureInfo.InvariantCulture, out degreeOfParallelism))
                        return false;
                }
                else
                {
                    return false;
                }
            }

            if (memorySize < 19 * 1024 || iterations < 2 || degreeOfParallelism < 1)
                return false;

            return true;
        }

        private sealed class ParsedHash
        {
            public required int MemorySize { get; init; }
            public required int Iterations { get; init; }
            public required int DegreeOfParallelism { get; init; }
            public required byte[] Salt { get; init; }
            public required byte[] Hash { get; init; }
        }
    }
}
