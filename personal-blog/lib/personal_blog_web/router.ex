defmodule PersonalBlogWeb.Router do
  use Phoenix.Router

  pipeline :browser do
    plug(:accepts, ["html"])
    plug(:fetch_session)
    plug(:fetch_live_flash)
    plug(:put_root_layout, {PersonalBlogWeb.Layouts, :root})
    plug(:protect_from_forgery)
    plug(:put_secure_browser_headers)
  end

  pipeline :api do
    plug(:accepts, ["json"])
  end

  scope "/", PersonalBlogWeb do
    pipe_through(:browser)

    live("/", BlogLive.Home, :home)
    live("/articles", BlogLive.Articles, :articles)
    live("/articles/:repo/:path", BlogLive.Article, :show)
    live("/portfolio", BlogLive.Portfolio, :portfolio)
    live("/portfolio/:repo/:path", BlogLive.PortfolioEntry, :show)
    live("/projects", BlogLive.Projects, :projects)
    live("/projects/:repo/:path", BlogLive.Project, :show)
    get("/healthz", HealthController, :check)
  end

  # Other scopes may use custom stacks.
  # scope "/api", PersonalBlogWeb do
  #   pipe_through :api
  # end

  # Enable LiveDashboard and Swoosh mailbox preview in development
  if Mix.env() in [:dev, :test] do
    import Phoenix.LiveDashboard.Router

    scope "/" do
      pipe_through(:browser)
      live_dashboard("/dashboard", metrics: PersonalBlogWeb.Telemetry)
    end
  end
end
