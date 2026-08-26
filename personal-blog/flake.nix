{
  description = "Personal Blog - Phoenix Microservice";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            elixir_1_17
            erlang_27
            elixir-ls
            git
            postgresql
            mix
          ];

          shellHook = ''
            mkdir -p .nix-mix .nix-hex
            export MIX_HOME=$PWD/.nix-mix
            export HEX_HOME=$PWD/.nix-hex
            export ERL_AFLAGS="-kernel shell_history enabled"
            export PATH="$MIX_HOME/bin:$PATH"
          '';
        };
      }
    );
}
