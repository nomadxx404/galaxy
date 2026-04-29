FROM mcr.microsoft.com/dotnet/sdk:10.0-noble AS build
WORKDIR /src

COPY ["auth-service.csproj", "./"]
RUN --mount=type=cache,id=nuget,target=/root/.nuget/packages \
    dotnet restore "auth-service.csproj" \
    -r linux-x64 \
    -p:PublishReadyToRun=true

COPY . .

RUN --mount=type=cache,id=nuget,target=/root/.nuget/packages \
    dotnet publish "auth-service.csproj" \
    -c Release \
    -o /app/publish \
    --no-restore \
    -r linux-x64 \
    --self-contained false \
    -p:PublishReadyToRun=true \
    -p:DebugType=None \
    -p:DebugSymbols=false

FROM mcr.microsoft.com/dotnet/aspnet:10.0-noble AS final
WORKDIR /app

COPY --from=build /app/publish .

ENV DOTNET_CLI_TELEMETRY_OPTOUT=1 \
    ASPNETCORE_HTTP_PORTS=8080 \
    PYTHONUNBUFFERED=1

USER app

EXPOSE 8080

ENTRYPOINT ["dotnet", "auth-service.dll"]
