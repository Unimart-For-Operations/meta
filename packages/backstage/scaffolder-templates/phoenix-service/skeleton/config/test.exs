import Config

config :${{ values.appName }}, ${{ values.moduleName }}Web.Endpoint,
  http: [ip: {127, 0, 0, 1}, port: 4002],
  server: false,
  secret_key_base: "test-only-secret-key-base-test-only-secret-key-base-test-only-sec"

config :logger, level: :warning
