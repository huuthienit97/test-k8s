const string Version = "polyglot-dotnet-3";

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/health", () => Results.Json(new
{
    status = "ok",
    service = "dotnet",
    stack = "dotnet",
    version = Version,
}));

app.MapGet("/hello", () => Results.Json(new
{
    message = "hello from dotnet",
    stack = "dotnet",
    version = Version,
}));

var port = Environment.GetEnvironmentVariable("PORT") ?? "8080";
app.Run($"http://0.0.0.0:{port}");
