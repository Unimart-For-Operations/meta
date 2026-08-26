defmodule PersonalBlogWeb.DocRenderer do
  @moduledoc """
  Renders Markdown to HTML with heading anchors and table of contents extraction.
  """

  def render_markdown(content) when is_binary(content) do
    content
    |> Earmark.as_html!()
    |> add_heading_anchors()
  end

  def render_markdown(_), do: ""

  defp add_heading_anchors(html) do
    html
    |> String.split(~r/(<h[1-6][^>]*>)/)
    |> Enum.reduce([], fn part, acc ->
      case Regex.match?(~r/^<h[1-6]/, part) do
        true ->
          # Extract the heading level
          level = String.slice(part, 2, 1)

          # Close the opening tag
          tag_close = String.replace(part, ~r/>$/, "")

          # Extract text from the next element
          case acc do
            [prev_text | rest] ->
              slug = String.downcase(String.replace(prev_text, ~r/[^\w\s-]/, ""))

              [
                "#{tag_close} id=\"#{slug}\">",
                prev_text
                | rest
              ]

            _ ->
              [part | acc]
          end

        false ->
          [part | acc]
      end
    end)
    |> Enum.reverse()
    |> Enum.join()
  end
end
