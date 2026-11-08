FROM mcr.microsoft.com/dotnet/sdk:10.0

ARG UID=1000
ARG GID=1000

ENV HOME=/home/dev \
    NUGET_PACKAGES=/home/dev/.nuget/packages

RUN mkdir -p $NUGET_PACKAGES /src/artifacts \
    && chown -R $UID:$GID /home/dev /src

USER $UID:$GID
WORKDIR /src

COPY --chown=$UID:$GID Finance.slnx Directory.Build.props dotnet-tools.json ./
COPY --chown=$UID:$GID src/Finance/Finance.csproj src/Finance/
COPY --chown=$UID:$GID tests/Finance.Tests/Finance.Tests.csproj tests/Finance.Tests/
RUN dotnet restore && dotnet tool restore

COPY --chown=$UID:$GID . .

EXPOSE 8002

CMD ["sleep", "infinity"]
