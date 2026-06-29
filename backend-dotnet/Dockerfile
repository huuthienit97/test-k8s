FROM mcr.microsoft.com/dotnet/sdk:8.0-alpine AS build
WORKDIR /src
COPY PolyglotApi.csproj .
RUN dotnet restore
COPY . .
RUN dotnet publish PolyglotApi.csproj -c Release -o /app /p:UseAppHost=false

FROM mcr.microsoft.com/dotnet/aspnet:8.0-alpine
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /app .
ENV ASPNETCORE_URLS=http://+:8080
ENV DOTNET_USE_POLLING_FILE_WATCHER=1
EXPOSE 8080
ENTRYPOINT ["dotnet", "PolyglotApi.dll"]
