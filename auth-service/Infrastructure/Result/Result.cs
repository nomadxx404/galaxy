namespace auth_service.Infrastructure
{
    public class Result<T>
    {
        public bool IsSuccess { get; }
        public int StatusCode { get; }
        public T? Data { get; }
        public string? Message { get; }

        private Result(bool isSuccess, int statusCode, string? message,T? data)
        {
            IsSuccess = isSuccess;
            StatusCode = statusCode;
            Message = message;
            Data = data;
        }

        public static Result<T> Success(int statusCode, string message, T data)
            => new(true, statusCode, message, data);
        public static Result<T> Success(int statusCode, string message)
            => new(true, statusCode, message, default);


        public static Result<T> Failure(int statusCode, string message)
            => new(false, statusCode, message, default);

        public static Result<T> Failure(int statusCode, string message, T data)
            => new(false, statusCode, message, data);
    }

    public class ApiMessageResponse
    {
        public bool Success { get; set; }
        public int StatusCode { get; set; }
        public string? Message { get; set; }
    }
}
