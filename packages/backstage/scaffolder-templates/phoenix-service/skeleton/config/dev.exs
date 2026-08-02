import Config

config :${{ values.appName }}, ${{ values.moduleName }}Web.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: ${{ values.port }}],
  check_origin: false,
  debug_errors: true,
  secret_key_base: "dev-only-secret-key-base-dev-only-secret-key-base-dev-only-secret"

config :logger, level: :debug
