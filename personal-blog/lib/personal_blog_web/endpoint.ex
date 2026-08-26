defmodule PersonalBlogWeb.Endpoint do
  use Phoenix.Endpoint, otp_app: :personal_blog

  # Serve at "/" the static files from "priv/static" directory.
  plug(Plug.Static,
    at: "/",
    from: :personal_blog,
    gzip: true,
    only: ~w(assets fonts images favicon.ico robots.txt),
    only_matching: ~w(~r/\.js(\.map)?$/ ~r/\.css(\.map)?$/ ~r/\.wasm$/)
  )

  # Code reloading can be explicitly enabled under the :code_reloader
  # configuration of your endpoint.
  if code_reloading? do
    plug(Phoenix.CodeReloader)
  end

  plug(Phoenix.LiveDashboard.RequestLogger,
    param_key: "request_logger",
    cookie_key: "request_logger"
  )

  plug(Plug.RequestId)
  plug(Plug.Parsers,
    parsers: [:urlencoded, :multipart, :json],
    pass: ["*/*"],
    json_decoder: Jason
  )

  plug(Plug.MethodOverride)
  plug(Plug.Head)
  plug(Plug.Session,
    store: :cookie,
    key: "_personal_blog_key",
    signing_salt: "personal_blog_salt"
  )

  plug(PersonalBlogWeb.Router)

  def code_reloading? do
    Application.get_env(:personal_blog, :code_reloading, Mix.env() != :prod)
  end
end
