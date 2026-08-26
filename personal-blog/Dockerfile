FROM hexpm/elixir:1.17.0-erlang-27.0.1-alpine-3.20 as builder

# Install build dependencies
RUN apk add --no-cache build-base git

WORKDIR /app

# Install hex and rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Copy mix files
COPY mix.exs mix.lock ./

# Install dependencies
RUN mix deps.get --only prod && \
    mix deps.compile

# Copy source code
COPY lib lib
COPY config config
COPY priv priv
COPY assets assets

# Compile assets
RUN mix tailwind.install --if-missing && \
    mix assets.deploy

# Compile the application
RUN mix compile && \
    mix release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openssl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create app user
RUN useradd -m -s /bin/bash app

WORKDIR /home/app

# Copy release from builder
COPY --from=builder --chown=app:app /app/_build/prod/rel/personal_blog ./

# Set environment
ENV MIX_ENV=prod

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/home/app/bin/personal_blog", "eval", "IO.puts(:ok)"]

# Ulimit for file descriptors
RUN ulimit -n 65536

# Run as non-root
USER app

EXPOSE 4000

# Start the application
CMD ["bin/personal_blog", "start"]
