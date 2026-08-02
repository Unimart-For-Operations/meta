{
  description = "${{ values.name }} — ${{ values.description }}";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          packages =
            with pkgs;
            [
              elixir
              erlang
              elixir-ls
            ]
            ++ lib.optionals stdenv.isLinux [ inotify-tools ]
            ++ lib.optionals stdenv.isDarwin [ fswatch ];

          shellHook = ''
            export MIX_HOME="$PWD/.nix-mix"
            export HEX_HOME="$PWD/.nix-hex"
            export PATH="$MIX_HOME/bin:$HEX_HOME/bin:$PATH"
          '';
        };

        formatter = pkgs.nixfmt-rfc-style;
      }
    );
}
