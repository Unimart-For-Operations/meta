defmodule PersonalBlogWeb do
  def controller do
    quote do
      use Phoenix.Controller,
        namespace: PersonalBlogWeb

      import Plug.Conn
      import PersonalBlogWeb.Gettext

      unquote(verified_routes())
    end
  end

  def live_view do
    quote do
      use Phoenix.LiveView,
        layout: {PersonalBlogWeb.Layouts, :app}

      import PersonalBlogWeb.Gettext
      unquote(verified_routes())
    end
  end

  def live_component do
    quote do
      use Phoenix.LiveComponent

      import PersonalBlogWeb.Gettext
      unquote(verified_routes())
    end
  end

  def router do
    quote do
      use Phoenix.Router

      import Plug.Conn
      import Phoenix.Controller
      import Phoenix.LiveView.Router
    end
  end

  def channel do
    quote do
      use Phoenix.Channel
      import PersonalBlogWeb.Gettext
    end
  end

  defp verified_routes do
    quote do
      use Phoenix.VerifiedRoutes,
        endpoint: PersonalBlogWeb.Endpoint,
        router: PersonalBlogWeb.Router
    end
  end

  @doc """
  When used, dispatch to the appropriate controller/view/etc.
  """
  defmacro __using__(which) when is_atom(which) do
    apply(__MODULE__, which, [])
  end
end
