defmodule PersonalBlog.Gitea do
  @moduledoc """
  Read-only Gitea API client for discovering and fetching content.
  
  Discovers content patterns:
  - portfolio/ - portfolio projects and experience
  - articles/ - blog articles and writings
  - projects/ - project showcases
  - README.md - main bio/intro
  - AGENTS.md - technical deep-dives
  """

  require Logger

  def gitea_url, do: System.get_env("GITEA_URL", "http://my-gitea-http.gitea.svc.cluster.local:3000")
  def gitea_token, do: System.get_env("GITEA_TOKEN", "")
  def gitea_user, do: System.get_env("GITEA_USER", "")

  # ===== Repo Discovery =====

  def list_repos do
    cache_key = "repos:list"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, repos} ->
        {:ok, repos}

      :miss ->
        case fetch_repos() do
          {:ok, repos} ->
            PersonalBlog.Cache.put(cache_key, repos)
            {:ok, repos}

          error ->
            error
        end
    end
  end

  defp fetch_repos do
    url = "#{gitea_url()}/api/v1/user/repos?limit=100"

    case request(:get, url) do
      {:ok, repos} ->
        filtered = Enum.filter(repos, &is_content_repo?/1)
        {:ok, filtered}

      error ->
        Logger.error("Failed to fetch repos from Gitea: #{inspect(error)}")
        {:error, :fetch_failed}
    end
  end

  defp is_content_repo?(repo) do
    # Only include repos with portfolio/, articles/, projects/, or docs
    repo["name"] not in [".", ".."] && not repo.get("archived", false)
  end

  # ===== Portfolio Discovery =====

  def list_portfolio_entries do
    case list_repos() do
      {:ok, repos} ->
        entries =
          repos
          |> Enum.flat_map(&list_files(&1["name"], "portfolio"))
          |> Enum.sort_by(& &1[:name])

        {:ok, entries}

      error ->
        error
    end
  end

  def get_portfolio_entry(repo, path) do
    cache_key = "portfolio:#{repo}:#{path}"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, content} ->
        {:ok, content}

      :miss ->
        case fetch_file(repo, "portfolio/#{path}") do
          {:ok, content} ->
            PersonalBlog.Cache.put(cache_key, content)
            {:ok, content}

          error ->
            error
        end
    end
  end

  # ===== Articles Discovery =====

  def list_articles do
    case list_repos() do
      {:ok, repos} ->
        articles =
          repos
          |> Enum.flat_map(&list_files(&1["name"], "articles"))
          |> Enum.sort_by(& &1[:modified], :desc)

        {:ok, articles}

      error ->
        error
    end
  end

  def get_article(repo, path) do
    cache_key = "article:#{repo}:#{path}"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, content} ->
        {:ok, content}

      :miss ->
        case fetch_file(repo, "articles/#{path}") do
          {:ok, content} ->
            PersonalBlog.Cache.put(cache_key, content)
            {:ok, content}

          error ->
            error
        end
    end
  end

  # ===== Projects Discovery =====

  def list_projects do
    case list_repos() do
      {:ok, repos} ->
        projects =
          repos
          |> Enum.flat_map(&list_files(&1["name"], "projects"))
          |> Enum.sort_by(& &1[:name])

        {:ok, projects}

      error ->
        error
    end
  end

  def get_project(repo, path) do
    cache_key = "project:#{repo}:#{path}"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, content} ->
        {:ok, content}

      :miss ->
        case fetch_file(repo, "projects/#{path}") do
          {:ok, content} ->
            PersonalBlog.Cache.put(cache_key, content)
            {:ok, content}

          error ->
            error
        end
    end
  end

  # ===== README Discovery =====

  def get_bio do
    cache_key = "bio:readme"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, content} ->
        {:ok, content}

      :miss ->
        user = gitea_user()

        if user == "" do
          {:error, :no_user_configured}
        else
          case fetch_file(user, "README.md") do
            {:ok, content} ->
              PersonalBlog.Cache.put(cache_key, content)
              {:ok, content}

            error ->
              error
          end
        end
    end
  end

  # ===== Internal Helpers =====

  defp list_files(repo, dir) do
    cache_key = "files:#{repo}:#{dir}"

    case PersonalBlog.Cache.get(cache_key) do
      {:ok, files} ->
        {:ok, files}

      :miss ->
        case fetch_tree(repo, dir) do
          {:ok, files} ->
            filtered =
              files
              |> Enum.filter(fn f -> String.ends_with?(f["name"], ".md") end)
              |> Enum.map(fn f ->
                %{
                  name: String.replace_suffix(f["name"], ".md", ""),
                  repo: repo,
                  path: f["path"],
                  modified: f["commit"]["committer"]["date"]
                }
              end)

            PersonalBlog.Cache.put(cache_key, filtered)
            filtered

          _error ->
            []
        end
    end
  end

  defp fetch_tree(repo, dir) do
    url =
      "#{gitea_url()}/api/v1/repos/#{gitea_user()}/#{repo}/contents/#{dir}?ref=main"

    case request(:get, url) do
      {:ok, items} when is_list(items) ->
        {:ok, items}

      _error ->
        {:error, :not_found}
    end
  end

  defp fetch_file(repo, path) do
    url =
      "#{gitea_url()}/api/v1/repos/#{gitea_user()}/#{repo}/contents/#{path}?ref=main"

    case request(:get, url) do
      {:ok, response} when is_map(response) ->
        case Base.decode64(response["content"]) do
          {:ok, content} ->
            {:ok, content}

          :error ->
            {:error, :decode_failed}
        end

      _error ->
        {:error, :not_found}
    end
  end

  # ===== HTTP Utilities =====

  defp request(method, url) do
    headers = []

    headers =
      if gitea_token() != "" do
        [{"Authorization", "token #{gitea_token()}"} | headers]
      else
        headers
      end

    try do
      response =
        Req.request!(
          method,
          url,
          headers: headers,
          receive_timeout: 5000
        )

      if response.status == 200 do
        {:ok, response.body}
      else
        {:error, :http_error}
      end
    rescue
      e ->
        Logger.error("Gitea request failed: #{inspect(e)}")
        {:error, :request_failed}
    end
  end
end
