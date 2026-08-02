defmodule ${{ values.moduleName }}Web.HealthController do
  use Phoenix.Controller, formats: [:json]

  def index(conn, _params), do: text(conn, "ok")

  def hello(conn, _params) do
    json(conn, %{service: "${{ values.name }}", message: "hello from Phoenix"})
  end
end
