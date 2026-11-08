FROM mcr.microsoft.com/dotnet/sdk:10.0

WORKDIR /src

COPY *.csproj ./
RUN dotnet restore

COPY . .

EXPOSE 8002

CMD ["sleep", "infinity"]
