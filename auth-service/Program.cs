using auth_service.Infrastructure;
using auth_service.Infrastructure.Middleware;
using DotNetEnv;
using Scalar.AspNetCore;

Env.Load();

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddControllers();
builder.Services.AddOpenApi();
builder.Services.AddHttpContextAccessor();

builder.Services.AddInfrastructure();

var app = builder.Build();

app.UseMiddleware<GlobalExceptionMiddleware>();
app.UseMiddleware<AuthContextMiddleware>();

app.MapOpenApi();   
app.MapScalarApiReference();
app.MapGet("/", () => Results.Redirect("/scalar"));

//if (app.Environment.IsDevelopment())
//{
//}

app.UseHttpsRedirection();

app.UseAuthorization();

app.MapControllers();

app.Run();
