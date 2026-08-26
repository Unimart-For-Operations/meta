import Config

# For production, don't include debug information in exception pagelets.
config :personal_blog, PersonalBlogWeb.Endpoint,
  cache_static_manifest: "priv/static/cache_manifest.json",
  force_ssl: false

# Do not print debug messages in production
config :logger, level: :info
