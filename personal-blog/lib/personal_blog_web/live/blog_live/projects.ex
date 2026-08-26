defmodule PersonalBlogWeb.BlogLive.Projects do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(_params, _session, socket) do
    case PersonalBlog.Gitea.list_projects() do
      {:ok, projects} ->
        {:ok, assign(socket, :projects, projects)}

      {:error, _reason} ->
        {:ok, assign(socket, :projects, [])}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div>
      <h1 class="text-4xl font-bold mb-8">Projects</h1>

      <%= if Enum.empty?(@projects) do %>
        <p class="text-slate-600 dark:text-slate-400">No projects found.</p>
      <% else %>
        <div class="space-y-6">
          <%= for project <- @projects do %>
            <a href={~p"/projects/#{project.repo}/#{project.path}"} class="block p-6 border border-slate-200 dark:border-slate-700 rounded-lg hover:shadow-lg transition">
              <h2 class="text-2xl font-bold mb-2"><%= project.name %></h2>
            </a>
          <% end %>
        </div>
      <% end %>
    </div>
    """
  end
end

defmodule PersonalBlogWeb.BlogLive.Project do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(%{"repo" => repo, "path" => path}, _session, socket) do
    case PersonalBlog.Gitea.get_project(repo, path) do
      {:ok, content} ->
        {:ok, assign(socket, repo: repo, path: path, content: content)}

      {:error, _reason} ->
        {:ok, assign(socket, repo: repo, path: path, content: "Project not found")}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="prose dark:prose-invert max-w-none">
      <div class="mb-8">
        <a href={~p"/projects"} class="text-blue-600 dark:text-blue-400 hover:underline">← Back to Projects</a>
      </div>
      <h1><%= @path %></h1>
      <%= raw(PersonalBlogWeb.DocRenderer.render_markdown(@content)) %>
    </div>
    """
  end
end
