import Config

if config_env() == :prod do
  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise "environment variable SECRET_KEY_BASE is missing"

  host = System.get_env("PHX_HOST") || "${{ values.name }}.cnoe.localtest.me"
  port = String.to_integer(System.get_env("PORT") || "${{ values.port }}")

  config :${{ values.appName }}, ${{ values.moduleName }}Web.Endpoint,
    url: [host: host, port: 443, scheme: "https"],
    http: [ip: {0, 0, 0, 0}, port: port],
    check_origin: false,
    secret_key_base: secret_key_base,
    server: true
end
