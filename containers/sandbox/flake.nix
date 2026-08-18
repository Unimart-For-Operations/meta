{
  description = "Bare-minimum TTY toolset for the nix-sandbox image";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      profiles = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        pkgs.buildEnv {
          name = "tty-profile";
          paths = with pkgs; [
            bat
            eza
            fd
            fzf
            git
            neovim
            openssh
            ripgrep
            starship
            tmux
            yazi
            zoxide
            zsh
          ];
        }
      );
    in
    {
      packages = forAllSystems (system: {
        default = profiles.${system};
        ttyProfile = profiles.${system};
      });
    };
}
