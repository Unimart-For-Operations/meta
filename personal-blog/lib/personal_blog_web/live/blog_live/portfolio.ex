defmodule PersonalBlogWeb.BlogLive.Portfolio do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(_params, _session, socket) do
    case PersonalBlog.Gitea.list_portfolio_entries() do
      {:ok, entries} ->
        {:ok, assign(socket, :entries, entries)}

      {:error, _reason} ->
        {:ok, assign(socket, :entries, [])}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div>
      <h1 class="text-4xl font-bold mb-8">Portfolio</h1>

      <%= if Enum.empty?(@entries) do %>
        <p class="text-slate-600 dark:text-slate-400">No portfolio entries found.</p>
      <% else %>
        <div class="space-y-6">
          <%= for entry <- @entries do %>
            <a href={~p"/portfolio/#{entry.repo}/#{entry.path}"} class="block p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
              <h2 class="text-2xl font-bold mb-2"><%= entry.name %></h2>
            </a>
          <% end %>
        </div>
      <% end %>
    </div>
    """
  end
end

defmodule PersonalBlogWeb.BlogLive.PortfolioEntry do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(%{"repo" => repo, "path" => path}, _session, socket) do
    case PersonalBlog.Gitea.get_portfolio_entry(repo, path) do
      {:ok, content} ->
        {:ok, assign(socket, repo: repo, path: path, content: content)}

      {:error, _reason} ->
        {:ok, assign(socket, repo: repo, path: path, content: "Entry not found")}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="prose dark:prose-invert max-w-none">
      <div class="mb-8">
        <a href={~p"/portfolio"} class="text-blue-600 dark:text-blue-400 hover:underline">← Back to Portfolio</a>
      </div>
      <h1><%= @path %></h1>
      <%= raw(PersonalBlogWeb.DocRenderer.render_markdown(@content)) %>
    </div>
    """
  end
end
