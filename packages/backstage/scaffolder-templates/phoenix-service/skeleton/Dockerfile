# Build stage — official hexpm image pins Elixir + OTP
FROM hexpm/elixir:1.17.3-erlang-27.1.2-debian-bookworm-20241111-slim AS builder

RUN apt-get update -y && apt-get install -y build-essential git ca-certificates \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV MIX_ENV=prod

RUN mix local.hex --force && mix local.rebar --force

COPY mix.exs mix.lock* ./
RUN mix deps.get --only prod && mix deps.compile

COPY config config
COPY lib lib
RUN mix compile && mix release

# Runtime stage
FROM debian:bookworm-slim AS runtime

RUN apt-get update -y \
    && apt-get install -y libstdc++6 openssl libncurses6 locales ca-certificates \
    && apt-get clean && rm -rf /var/lib/apt/lists/* \
    && sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen && locale-gen

ENV LANG=en_US.UTF-8 LANGUAGE=en_US:en LC_ALL=en_US.UTF-8 MIX_ENV=prod

WORKDIR /app
RUN useradd --create-home app && chown app /app
USER app

COPY --from=builder --chown=app:app /app/_build/prod/rel/${{ values.appName }} ./

EXPOSE ${{ values.port }}
CMD ["/app/bin/${{ values.appName }}", "start"]
