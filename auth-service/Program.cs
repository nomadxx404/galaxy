using auth_service.Infrastructure;
using auth_service.Infrastructure.Middleware;
using DotNetEnv;
using Microsoft.OpenApi;
using Scalar.AspNetCore;

Env.Load();

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddControllers();

builder.Services.AddOpenApi(options =>
{
    options.AddDocumentTransformer((document, context, cancellationToken) =>
    {
        document.Servers = new List<OpenApiServer>
        {
            new() { Url = "http://localhost:8000" }
        };
        return Task.CompletedTask;
    });
});

builder.Services.AddHttpContextAccessor();

builder.Services.AddInfrastructure();

builder.Services.AddHealthChecks();

var app = builder.Build();

app.UseMiddleware<GlobalExceptionMiddleware>();

app.MapHealthChecks("/health");

app.MapOpenApi();
app.MapScalarApiReference("/scalar");

//if (app.Environment.IsDevelopment())
//{
//}

app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();

app.Run();
