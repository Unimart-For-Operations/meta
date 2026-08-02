import Config

config :${{ values.appName }}, ${{ values.moduleName }}Web.Endpoint,
  url: [host: "localhost"],
  adapter: Bandit.PhoenixAdapter,
  render_errors: [formats: [html: ${{ values.moduleName }}Web.ErrorHTML], layout: false],
  pubsub_server: ${{ values.moduleName }}.PubSub

config :phoenix, :json_library, Jason

import_config "#{config_env()}.exs"
