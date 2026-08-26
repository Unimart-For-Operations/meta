import Config

# For development, we disable any cache and enable
# debugging and code reloading.
#
# The watchers configuration allows you to run other
# applications or custom logic as part of your supervision tree
# during development with the ability to update the watched
# source without restarting your server.
#
# By default, the watchers includes anything in the `assets` directory
# using the `esbuild` watched mode as the default.
if config_env() == :dev do
  config :personal_blog, PersonalBlogWeb.Endpoint,
    # Binding to loopback ipv4 address prevents access from other machines.
    # Change to `ip: {0, 0, 0, 0}` to allow access from other machines.
    http: [ip: {127, 0, 0, 1}, port: 4000],
    check_origin: false,
    code_reloader: true,
    debug_errors: true,
    secret_key_base: "dev_secret_key_base_change_me_in_production",
    watchers: [
      tailwind: {Tailwind, :install_and_run, [:default, ~w(--watch)]}
    ],
    live_reload: [
      patterns: [
        ~r"priv/static/.*(js|css|png|jpeg|jpg|gif|svg)$",
        ~r"priv/gettext/.*(po)$",
        ~r"lib/personal_blog_web/(controllers|live|components)/.*(ex|heex)$"
      ]
    ]

  config :logger, level: :debug
end
