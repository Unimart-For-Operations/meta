defmodule PersonalBlogWeb.Layouts do
  use PersonalBlogWeb, :verified_routes
  import Phoenix.HTML

  embed_templates "layouts/*"
end
