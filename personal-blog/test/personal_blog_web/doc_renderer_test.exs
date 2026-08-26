defmodule PersonalBlogWeb.DocRendererTest do
  use ExUnit.Case

  alias PersonalBlogWeb.DocRenderer

  test "renders markdown to HTML" do
    markdown = "# Hello World\n\nThis is a test."
    result = DocRenderer.render_markdown(markdown)
    assert result =~ "<h1"
    assert result =~ "Hello World"
    assert result =~ "This is a test"
  end

  test "handles empty markdown" do
    result = DocRenderer.render_markdown("")
    assert result == ""
  end

  test "handles nil gracefully" do
    result = DocRenderer.render_markdown(nil)
    assert result == ""
  end

  test "adds heading anchors" do
    markdown = "# My Heading"
    result = DocRenderer.render_markdown(markdown)
    assert result =~ "id="
  end
end
