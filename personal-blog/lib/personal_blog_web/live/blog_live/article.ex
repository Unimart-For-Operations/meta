defmodule PersonalBlogWeb.BlogLive.Article do
  use PersonalBlogWeb, :live_view

  @impl true
  def mount(%{"repo" => repo, "path" => path}, _session, socket) do
    case PersonalBlog.Gitea.get_article(repo, path) do
      {:ok, content} ->
        {:ok, assign(socket, repo: repo, path: path, content: content)}

      {:error, _reason} ->
        {:ok, assign(socket, repo: repo, path: path, content: "Article not found")}
    end
  end

  @impl true
  def render(assigns) do
    ~H"""
    <div class="prose dark:prose-invert max-w-none">
      <div class="mb-8">
        <a href={~p"/articles"} class="text-blue-600 dark:text-blue-400 hover:underline">← Back to Articles</a>
      </div>
      <h1><%= @path %></h1>
      <%= raw(PersonalBlogWeb.DocRenderer.render_markdown(@content)) %>
    </div>
    """
  end
end
