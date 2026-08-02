defmodule ${{ values.moduleName }}.MixProject do
  use Mix.Project

  def project do
    [
      app: :${{ values.appName }},
      version: "0.1.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      releases: [
        ${{ values.appName }}: [include_executables_for: [:unix]]
      ]
    ]
  end

  def application do
    [
      mod: {${{ values.moduleName }}.Application, []},
      extra_applications: [:logger, :runtime_tools]
    ]
  end

  defp deps do
    [
      {:phoenix, "~> 1.7.14"},
      {:phoenix_html, "~> 4.1"},
      {:bandit, "~> 1.5"},
      {:jason, "~> 1.4"}
    ]
  end
end
