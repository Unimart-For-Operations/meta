{
  description = "unimart — unified CLI for the idpbuilder organization";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    {
      # Overlay for consumers (e.g. cmdr) to get pkgs.unimart
      overlays.default = final: _prev: {
        unimart = self.packages.${final.system}.unimart;
      };
    }
    // flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages = {
          unimart = pkgs.buildGoModule {
            pname = "unimart";
            version = version;
            src = ./.;
            vendorHash = "sha256-DWBIYmPJADeC8HD5bKw7H0c2/Xu+Jp/UkLO4wr5L1Jk=";

            ldflags = [
              "-s"
              "-w"
              "-X github.com/Unimart-For-Operations/meta/cmd.Version=${version}"
              "-X github.com/Unimart-For-Operations/meta/cmd.GitCommit=${version}"
              "-X github.com/Unimart-For-Operations/meta/cmd.BuildDate=1970-01-01T00:00:00Z"
            ];

            postInstall = ''
              mv $out/bin/meta $out/bin/unimart
            '';

            meta = with pkgs.lib; {
              description = "Unified CLI for the idpbuilder organization";
              homepage = "https://github.com/Unimart-For-Operations/meta";
              license = licenses.asl20;
              mainProgram = "unimart";
            };
          };

          default = self.packages.${system}.unimart;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            goreleaser
          ];
        };

        formatter = pkgs.writeShellApplication {
          name = "meta-nixfmt";
          runtimeInputs = [ pkgs.nixfmt-rfc-style ];
          text = ''
            check=0
            if [ "''${1:-}" = "--check" ]; then
              check=1
              shift
            fi

            mapfile -t files < <(git ls-files '*.nix')
            if [ "''${#files[@]}" -eq 0 ]; then
              exit 0
            fi

            if [ "$check" -eq 1 ]; then
              exec nixfmt --check "''${files[@]}"
            fi
            exec nixfmt "''${files[@]}"
          '';
        };
      }
    );
}
