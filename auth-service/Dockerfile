FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

COPY ["auth-service.csproj", "./"]

RUN dotnet restore "auth-service.csproj"

COPY . .

RUN dotnet publish "auth-service.csproj" -c Release -o /app/publish --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app
COPY --from=build /app/publish .

ENTRYPOINT ["dotnet", "auth-service.dll"]