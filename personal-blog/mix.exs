defmodule PersonalBlog.MixProject do
  use Mix.Project

  def project do
    [
      app: :personal_blog,
      version: "0.1.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      aliases: aliases(),
      deps: deps(),
      releases: [
        personal_blog: [
          include_executables_for: [:unix],
          applications: [runtime_tools: :permanent]
        ]
      ]
    ]
  end

  def application do
    [
      extra_applications: [:logger],
      mod: {PersonalBlog.Application, []}
    ]
  end

  defp deps do
    [
      # Web framework
      {:phoenix, "~> 1.7.14"},
      {:phoenix_live_view, "~> 1.0"},
      {:bandit, "~> 1.5"},

      # HTTP client
      {:req, "~> 0.5"},

      # Markdown rendering
      {:earmark, "~> 1.4"},

      # JSON encoding
      {:jason, "~> 1.4"},

      # CSS tooling
      {:tailwind, "~> 0.2"},

      # Development & Test
      {:phoenix_live_reload, "~> 1.4", only: :dev},
      {:ex_doc, "~> 0.34", only: :dev},
      {:dialyxir, "~> 1.4", only: [:dev, :test], runtime: false},
      {:credo, "~> 1.7", only: [:dev, :test], runtime: false}
    ]
  end

  defp aliases do
    [
      setup: ["deps.get", "assets.setup"],
      "assets.setup": ["tailwind.install --if-missing"],
      "assets.build": ["tailwind default"],
      "assets.deploy": ["tailwind default --minify"],
      test: ["ecto.create --quiet", "ecto.migrate --quiet", "test"]
    ]
  end
end
