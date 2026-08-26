defmodule PersonalBlogWeb.BlogLive.Home do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(_params, _session, socket) do
    case PersonalBlog.Gitea.get_bio() do
      {:ok, bio} ->
        {:ok, assign(socket, :bio, bio)}

      {:error, _reason} ->
        {:ok, assign(socket, :bio, "Welcome to my personal blog!")}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="prose dark:prose-invert max-w-none">
      <h1>Welcome</h1>
      <div class="mb-8">
        <%= raw(PersonalBlogWeb.DocRenderer.render_markdown(@bio)) %>
      </div>

      <div class="grid grid-cols-3 gap-6 my-12">
        <a href={~p"/portfolio"} class="p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
          <h3 class="text-xl font-bold mb-2">Portfolio</h3>
          <p class="text-slate-600 dark:text-slate-400">Experience and accomplishments</p>
        </a>

        <a href={~p"/projects"} class="p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
          <h3 class="text-xl font-bold mb-2">Projects</h3>
          <p class="text-slate-600 dark:text-slate-400">Showcase of work and ideas</p>
        </a>

        <a href={~p"/articles"} class="p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
          <h3 class="text-xl font-bold mb-2">Articles</h3>
          <p class="text-slate-600 dark:text-slate-400">Thoughts and technical writings</p>
        </a>
      </div>
    </div>
    """
  end
end
