defmodule PersonalBlog.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      {PersonalBlog.Cache, []},
      PersonalBlogWeb.Telemetry,
      {DNSCluster, query: Application.get_env(:personal_blog, :dns_cluster_query) || :ignore},
      {Bandit, bandit_opts()}
    ]

    opts = [strategy: :one_for_one, name: PersonalBlog.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp bandit_opts do
    [
      scheme: :http,
      plug: PersonalBlogWeb.Router,
      port: String.to_integer(System.get_env("PORT", "4000")),
      thousand_island_options: [
        transport_options: [
          max_connections: 10_000,
          shutdown_timeout: 1000
        ]
      ]
    ]
  end
end
