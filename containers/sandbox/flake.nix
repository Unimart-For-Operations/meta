{
  description = "Nix sandbox — 1:1 parity with cmdr tty + tui modules";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    nixvim = {
      url = "github:nix-community/nixvim";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      home-manager,
      nixvim,
      ...
    }@inputs:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;

      # Synthetic host metadata for the sandbox — mirrors the schema
      # from cmdr/home/02-hosts/<distro>/<host>/meta.nix.
      sandboxMeta = {
        description = "nix-sandbox — containerised TTY workstation";
        system = "x86_64-linux";
        username = "root";
        homeDirectory = "/root";
        gitName = "Sandbox User";
        gitEmail = "sandbox@cnoe.localtest.me";
        role = "tty-engineer";
        capabilities = [
          "baseline"
          "terminal-dev"
        ];
        features = [
          "tty"
          "tui"
        ];
      };

      # cmdr/home is copied into the flake root by the Dockerfile so all
      # paths stay inside the flake source tree (pure evaluation).
      cmdr = ./cmdr-home;

      mkSandboxConfig =
        system:
        home-manager.lib.homeManagerConfiguration {
          pkgs = import nixpkgs {
            inherit system;
            config.allowUnfree = true;
          };

          modules = [
            # ── Core baseline ──────────────────────────────────────
            (cmdr + /03-features/base.nix)
            (cmdr + /01-platforms/linux.nix)

            # ── TTY feature (full parity) ──────────────────────────
            (cmdr + /03-features/tty.nix)

            # ── TUI graduated modules (no incubating) ──────────────
            (cmdr + /04-modules/tui/graduated/lazygit)
            (cmdr + /04-modules/tui/graduated/k9s)

            # ── Cherry-picked CLI modules ──────────────────────────
            (cmdr + /04-modules/cli/graduated/atuin)
            (cmdr + /04-modules/cli/graduated/direnv)
            (cmdr + /04-modules/cli/graduated/fonts)
            (cmdr + /04-modules/cli/graduated/opencode)

            # ── Container-specific overrides ───────────────────────
            (
              { lib, ... }:
              {
                home = {
                  username = sandboxMeta.username;
                  homeDirectory = sandboxMeta.homeDirectory;
                };

                # Suppress news — no interactive terminal during build
                news.display = "silent";

                # AstroNvim's lazy.nvim clones plugins over HTTPS from GitHub.
                # The org-wide git module rewrites https://github.com/ ->
                # git@github.com: (SSH), which fails in the sandbox — no SSH key
                # is deployed and outbound SSH is typically blocked. Force an
                # empty `url` section so lazy.nvim clones over HTTPS at runtime.
                programs.git.settings.url = lib.mkForce { };
              }
            )
          ];

          extraSpecialArgs = {
            inherit inputs;
            hostName = "sandbox";
            hostMeta = sandboxMeta;
          };
        };
    in
    {
      # The activation package is what `home-manager switch` would
      # install.  The Dockerfile builds it and runs the activation
      # script to populate /root with the fully-configured dotfiles.
      packages = forAllSystems (system: {
        default = (mkSandboxConfig system).activationPackage;
        activationPackage = (mkSandboxConfig system).activationPackage;
      });
    };
}
