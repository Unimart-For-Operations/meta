import Config

# Configure the logger
config :logger, level: :info

# Use Bandit for HTTP adapter
config :personal_blog,
  ecto_repos: []

config :personal_blog_web, PersonalBlogWeb.Endpoint,
  adapter: Bandit.PhoenixAdapter,
  render_errors: [
    view: PersonalBlogWeb.ErrorJSON,
    accepts: ~w(json),
    layout: false
  ],
  pubsub_server: PersonalBlog.PubSub,
  live_view: [
    signing_salt: "personal_blog_live_view_salt"
  ]

# Configures Tailwind
config :tailwind,
  version: "3.4.17",
  default: [
    args: ~w(
      --config=tailwind.config.js
      --input=assets/css/app.css
      --output=priv/static/assets/app.css
    ),
    cd: Path.expand("../", __DIR__)
  ]

# Gettext
config :gettext, :default_locale, "en"
