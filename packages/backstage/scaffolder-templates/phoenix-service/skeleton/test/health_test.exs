defmodule ${{ values.moduleName }}Web.HealthTest do
  use ExUnit.Case, async: true

  import Plug.Test

  test "GET /healthz returns ok" do
    conn = conn(:get, "/healthz") |> ${{ values.moduleName }}Web.Router.call([])
    assert conn.status == 200
    assert conn.resp_body == "ok"
  end
end
