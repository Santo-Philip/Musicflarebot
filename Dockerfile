FROM golang:1.26-bookworm AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    zlib1g-dev \
    unzip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN set -eux; \
    RELEASE_URL="https://api.github.com/repos/pytgcalls/ntgcalls/releases/tags/v2.2.1-beta02"; \
    ASSET="ntgcalls.linux-x86_64-static_libs.zip"; \
    DL_URL=$(curl -sSL "$RELEASE_URL" | grep -oP '"browser_download_url": "\K[^"]*(?=")' | grep -F "$ASSET" | head -1); \
    if [ -z "$DL_URL" ]; then echo "no asset found"; exit 1; fi; \
    curl -sL -o ntgcalls.zip "$DL_URL"; \
    unzip -qo ntgcalls.zip -d ntgcalls_tmp; \
    for f in $(find ntgcalls_tmp -name 'ntgcalls.h'); do cp "$f" src/vc/ntgcalls/ntgcalls.h; done; \
    for f in $(find ntgcalls_tmp -name 'libntgcalls.*'); do cp "$f" src/vc/; done; \
    rm -rf ntgcalls.zip ntgcalls_tmp

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/bot/

FROM debian:12-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    wget \
    unzip \
    curl \
    lsb-release \
    ca-certificates \
    libatomic1 \
    && rm -rf /var/lib/apt/lists/*

RUN wget -O /usr/local/bin/yt-dlp \
    https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp_linux \
    && chmod +x /usr/local/bin/yt-dlp

RUN curl -fsSL https://deno.land/install.sh | sh \
    && export DENO_INSTALL="/opt/deno" \
    && export PATH="$DENO_INSTALL/bin:$PATH" \
    && mv /root/.deno /opt/deno \
    && ln -sf /opt/deno/bin/deno /usr/local/bin/deno

RUN groupadd -r app && useradd -r -g app -m -d /home/app app

ENV DENO_INSTALL="/opt/deno"
ENV PATH="${DENO_INSTALL}/bin:${PATH}"
ENV HOME="/home/app"

COPY --from=builder --chown=app:app /app/main /usr/local/bin/app

RUN chown -R app:app /opt/deno

USER app

WORKDIR /home/app
ENTRYPOINT ["app"]
