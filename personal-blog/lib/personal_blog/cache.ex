defmodule PersonalBlog.Cache do
  @moduledoc """
  ETS-backed TTL cache for Gitea API responses and content.
  Default TTL is 60 seconds, configurable via BLOG_CACHE_TTL env var.
  """

  use GenServer

  require Logger

  def start_link(opts) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @impl true
  def init(_opts) do
    :ets.new(:personal_blog_cache, [:set, :named_table, :protected])
    {:ok, %{ttl: get_ttl()}}
  end

  def get(key) do
    case :ets.lookup(:personal_blog_cache, key) do
      [{^key, value, expires_at}] ->
        if System.monotonic_time(:millisecond) < expires_at do
          {:ok, value}
        else
          :ets.delete(:personal_blog_cache, key)
          :miss
        end

      [] ->
        :miss
    end
  end

  def put(key, value) do
    ttl = get_ttl()
    expires_at = System.monotonic_time(:millisecond) + ttl

    :ets.insert(:personal_blog_cache, {key, value, expires_at})
    :ok
  end

  def delete(key) do
    :ets.delete(:personal_blog_cache, key)
    :ok
  end

  def flush do
    :ets.delete_all_objects(:personal_blog_cache)
    :ok
  end

  defp get_ttl do
    System.get_env("BLOG_CACHE_TTL", "60")
    |> String.to_integer()
    |> Kernel.*(1000)
  end
end
