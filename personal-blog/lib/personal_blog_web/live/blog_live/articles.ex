defmodule PersonalBlogWeb.BlogLive.Articles do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(_params, _session, socket) do
    case PersonalBlog.Gitea.list_articles() do
      {:ok, articles} ->
        {:ok, assign(socket, :articles, articles)}

      {:error, _reason} ->
        {:ok, assign(socket, :articles, [])}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div>
      <h1 class="text-4xl font-bold mb-8">Articles</h1>

      <%= if Enum.empty?(@articles) do %>
        <p class="text-slate-600 dark:text-slate-400">No articles found.</p>
      <% else %>
        <div class="space-y-6">
          <%= for article <- @articles do %>
            <a href={~p"/articles/#{article.repo}/#{article.path}"} class="block p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
              <h2 class="text-2xl font-bold mb-2"><%= article.name %></h2>
              <p class="text-slate-600 dark:text-slate-400"><%= Calendar.strftime(String.to_existing_atom(article.modified), "%B %d, %Y") %></p>
            </a>
          <% end %>
        </div>
      <% end %>
    </div>
    """
  end
end
