defmodule ${{ values.moduleName }}Web.Endpoint do
  use Phoenix.Endpoint, otp_app: :${{ values.appName }}

  plug Plug.RequestId
  plug Plug.Telemetry, event_prefix: [:phoenix, :endpoint]

  plug Plug.Parsers,
    parsers: [:urlencoded, :multipart, :json],
    pass: ["*/*"],
    json_decoder: Phoenix.json_library()

  plug Plug.MethodOverride
  plug Plug.Head
  plug ${{ values.moduleName }}Web.Router
end
