defmodule PersonalBlogWeb.Gettext do
  @moduledoc """
  A module providing Internationalization with a default locale of :en.

  By default, Gettext extracts messages from your views and controllers
  in a `.pot` file. This is often the source file used to generate a .po files
  under the `priv/gettext` directory.

  However, this behaviour can be disabled by explicitly setting
  `:extract` to `false` for the `:gettext` config for your application.
  """
  use Gettext, otp_app: :personal_blog
end
