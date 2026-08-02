defmodule ${{ values.moduleName }}.Application do
  @moduledoc false
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      {Phoenix.PubSub, name: ${{ values.moduleName }}.PubSub},
      ${{ values.moduleName }}Web.Endpoint
    ]

    Supervisor.start_link(children, strategy: :one_for_one, name: ${{ values.moduleName }}.Supervisor)
  end

  @impl true
  def config_change(changed, _new, removed) do
    ${{ values.moduleName }}Web.Endpoint.config_change(changed, removed)
    :ok
  end
end
