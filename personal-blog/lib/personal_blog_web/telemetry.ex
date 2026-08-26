defmodule PersonalBlogWeb.Telemetry do
  use Supervisor
  import Telemetry.Metrics

  def start_link(arg) do
    Supervisor.start_link(__MODULE__, arg, name: __MODULE__)
  end

  @impl true
  def init(_arg) do
    children = [
      # Telemetry poller will execute the given period measurements
      # every 10_000ms. Learn more here: https://hexdocs.pm/telemetry_metrics
      {:telemetry_poller, Telemetry.Poller.periodic_measurements(__MODULE__.periodic_measurements())}
    ]

    Supervisor.init(children, strategy: :one_for_one)
  end

  def periodic_measurements do
    [
      # VM Metrics
      {Unit.BytesPerSecond, event_name: [:vm, :memory_total], measurement: :total},
      {Unit.Percent, event_name: [:vm, :memory_atom], measurement: :atom}
    ]
  end
end
