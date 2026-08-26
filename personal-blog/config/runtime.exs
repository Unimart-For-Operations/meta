import Config

# Runtime configuration (production only)
if config_env() == :prod do
  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise """
      environment variable SECRET_KEY_BASE is missing.
      You can generate one by calling: mix phx.gen.secret
      """

  host = System.get_env("PHX_HOST") || "localhost"
  port = String.to_integer(System.get_env("PORT") || "4000")

  config :personal_blog, PersonalBlogWeb.Endpoint,
    url: [host: host, port: 443, scheme: "https"],
    http: [
      # Enable IPv6 and bind on all available addresses.
      # Set it to  {0, 0, 0, 0, 0, 0, 0, 1} for local network only access.
      # See the documentation on https://hexdocs.pm/bandit/Bandit.html for details about using IPv6 vs IPv4 and loopback vs public addresses.
      ip: {0, 0, 0, 0},
      port: port
    ],
    secret_key_base: secret_key_base,
    server: true

  # Configure Gitea client
  config :personal_blog,
    gitea_url: System.get_env("GITEA_URL", "http://my-gitea-http.gitea.svc.cluster.local:3000"),
    gitea_token: System.get_env("GITEA_TOKEN", ""),
    gitea_user: System.get_env("GITEA_USER", "")

  # Reduce memory usage in production
  config :logger, level: :info
end
