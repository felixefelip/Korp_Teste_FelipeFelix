FROM mcr.microsoft.com/dotnet/sdk:10.0

ARG UID=1000
ARG GID=1000

ENV HOME=/home/dev \
    NUGET_PACKAGES=/home/dev/.nuget/packages

RUN mkdir -p $NUGET_PACKAGES /src/bin /src/obj \
    && chown -R $UID:$GID /home/dev /src

USER $UID:$GID
WORKDIR /src

COPY --chown=$UID:$GID *.csproj ./
RUN dotnet restore

COPY --chown=$UID:$GID . .

EXPOSE 8002

CMD ["sleep", "infinity"]
