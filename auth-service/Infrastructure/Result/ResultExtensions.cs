using Microsoft.AspNetCore.Mvc;

namespace auth_service.Infrastructure
{
    public static class ResultExtensions
    {
        public static IActionResult ToActionResult<T>(this Result<T> result, ControllerBase controller)
        {
            return controller.StatusCode(result.StatusCode, new
            {
                success = result.IsSuccess,
                message = result.Message,
                data = result.Data,
                status = result.StatusCode
            });
        }
    }
}