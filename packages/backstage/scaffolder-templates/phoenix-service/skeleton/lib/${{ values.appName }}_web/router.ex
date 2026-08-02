defmodule ${{ values.moduleName }}Web.Router do
  use Phoenix.Router

  pipeline :api do
    plug :accepts, ["json"]
  end

  scope "/", ${{ values.moduleName }}Web do
    pipe_through :api

    get "/healthz", HealthController, :index
    get "/", HealthController, :hello
  end
end
