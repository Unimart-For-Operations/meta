{
  # Human-facing description shown by `nix flake show` and similar commands.
  # This is metadata only; it does not affect how the flake evaluates.
  description = "unimart — unified CLI for the idpbuilder organization";

  inputs = {
    # Primary package set. Using nixpkgs-unstable keeps developer tooling and
    # Go packaging support current, which matters for an actively developed CLI.
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

    # Helper library that expands one output definition across the common host
    # systems, avoiding duplicated package/devShell declarations per platform.
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    {
      # Export an overlay so downstream flakes can say `pkgs.unimart` after
      # importing this flake's overlay, instead of reaching into
      # `inputs.meta.packages.${system}.unimart` directly.
      #
      # `final` is the fully composed package set for the consumer's system.
      # We intentionally ignore `prev` because this overlay only adds a new
      # package; it does not override an existing derivation.
      overlays.default = final: _prev: {
        unimart = self.packages.${final.system}.unimart;
      };
    }
    # Merge the overlay-only attribute set above with a per-system attribute
    # set generated for all default platforms supported by flake-utils.
    #
    # This is the standard flake pattern for publishing both:
    # - top-level attributes such as overlays
    # - system-specific outputs such as packages and devShells
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        # Materialize the nixpkgs package set for the current target system
        # (for example x86_64-linux or aarch64-darwin).
        pkgs = nixpkgs.legacyPackages.${system};

        # Prefer a real git revision when available so `unimart version` can
        # report something meaningful from flake metadata. During dirty local
        # development, `dirtyShortRev` still gives a useful identifier. If the
        # source is not a git checkout at all, fall back to a stable dev label.
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages = {
          # Main installable artifact exposed as `.#unimart`.
          unimart = pkgs.buildGoModule {
            # `pname` determines the default derivation name; the compiled Go
            # binary name still comes from the module/package being built.
            pname = "unimart";
            version = version;

            # Build from the repository root so the main Go module, embedded
            # assets, and local `replace` directives all resolve correctly.
            src = ./.;

            # Fixed-output hash for the vendored dependency tree produced by
            # `buildGoModule`. Any dependency graph change requires updating
            # this hash, which keeps Nix builds reproducible.
            vendorHash = "sha256-uIQJbJpSbA5q2a/JIWIntIadolMbTb5vlfAlWgKITy8=";

            # Build only the root package. The nested idpbuilder module
            # (./idpbuilder) is imported via replace => ./idpbuilder and is
            # compiled transitively, so auto-discovering it as a subpackage
            # would mis-treat it as part of the main module.
            subPackages = [ "." ];

            # Unit tests (internal/repos) shell out to `git` (e.g. `git init`
            # for temp test repos); make it available during the check phase.
            nativeBuildInputs = [ pkgs.git ];

            # Strip debug symbols for a smaller binary, then inject version
            # metadata into the Cobra command package at link time.
            #
            # BuildDate is intentionally fixed so local rebuilds of the same
            # source revision remain reproducible instead of embedding wall time.
            ldflags = [
              "-s"
              "-w"
              "-X github.com/Unimart-For-Operations/meta/cmd.Version=${version}"
              "-X github.com/Unimart-For-Operations/meta/cmd.GitCommit=${version}"
              "-X github.com/Unimart-For-Operations/meta/cmd.BuildDate=1970-01-01T00:00:00Z"
            ];

            # `buildGoModule` names the produced executable after the module's
            # main package, which here would be `meta`. Rename it so the final
            # installed program matches the user-facing CLI name.
            postInstall = ''
              mv $out/bin/meta $out/bin/unimart
            '';

            # Package metadata used by Nix tooling and downstream consumers.
            meta = with pkgs.lib; {
              description = "Unified CLI for the idpbuilder organization";
              homepage = "https://github.com/Unimart-For-Operations/meta";
              license = licenses.asl20;
              mainProgram = "unimart";
            };
          };

          # Make `nix build` with no explicit package select the same artifact
          # as `nix build .#unimart` for this system.
          default = self.packages.${system}.unimart;
        };

        # Development shell for contributors working on the Go CLI itself.
        # This is intentionally small: only the core Go toolchain and release
        # tooling needed for everyday development are provided here.
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            goreleaser
          ];
        };

        # Expose a formatter app so `nix fmt` uses repo-specific behavior
        # instead of blindly formatting every `.nix` file in the tree.
        formatter = pkgs.writeShellApplication {
          name = "meta-nixfmt";

          # The generated shell script gets `nixfmt` on PATH from here.
          runtimeInputs = [ pkgs.nixfmt ];
          text = ''
            # Support an optional `--check` mode so the same formatter can be
            # used both for CI-style verification and in-place rewrites.
            check=0
            if [ "''${1:-}" = "--check" ]; then
              check=1
              shift
            fi

            # Exclude Backstage scaffolder templates: their *.nix skeletons
            # carry scaffolder placeholders (dollar-brace pairs) that are
            # invalid as Nix by design, so nixfmt cannot parse them.
            #
            # The doubled single-quote before shell interpolation syntax is
            # Nix shell escaping:
            # it prevents Nix from interpolating the shell expression at flake
            # evaluation time, leaving the expansion for bash to handle at
            # runtime.
            mapfile -t files < <(git ls-files '*.nix' | grep -v 'scaffolder-templates/')
            if [ "''${#files[@]}" -eq 0 ]; then
              exit 0
            fi

            # In check mode, nixfmt exits non-zero if any file differs.
            if [ "$check" -eq 1 ]; then
              exec nixfmt --check "''${files[@]}"
            fi

            # Otherwise rewrite the tracked Nix files in place.
            exec nixfmt "''${files[@]}"
          '';
        };
      }
    );
}
