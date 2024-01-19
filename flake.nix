{
  description = "A Nix-flake-based Go 1.17 development environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  inputs.flake-compat.url = "github:edolstra/flake-compat";
  inputs.flake-compat.flake = false;
  inputs.solc-bin.url = "github:EspressoSystems/nix-solc-bin";

  outputs = { flake-utils, nixpkgs, solc-bin, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        goVersion = 21; # Change this to update the whole stack
        overlays = [
          # solc-bin.overlays.default
          (final: prev: {
            go = prev."go_1_${toString goVersion}";
            # Overlaying nodejs here to ensure nodePackages use the desired
            # version of nodejs.
            nodejs = prev.nodejs_20;
            pnpm = prev.nodePackages.pnpm;
            yarn = prev.nodePackages.yarn;
          })
        ];

        pkgs = import nixpkgs {
          inherit overlays system;
        };
        # nixWithFlakes allows pre v2.4 nix installations to use
        # flake commands (like `nix flake update`)
        nixWithFlakes = pkgs.writeShellScriptBin "nix" ''
          exec ${pkgs.nixFlakes}/bin/nix --experimental-features "nix-command flakes" "$@"
        '';
        foundry = pkgs.callPackage ./foundry { solc-bin-src = solc-bin; };
        # Overriding version for go-packages is more complicated than it should be.
        # To use a different go ethereum version the `version`, `hash` and `vendorHash` need to be updated.
        #
        # 1. Run the following command for `hash` :
        #
        #     nix-shell -p nix-prefetch-github --run "nix-prefetch-github ethereum go-ethereum --nix --rev v1.13.4"
        #
        # 2. Replace the `vendorHash` with an empty string and run `nix develop`.
        # 3. Find the `vendorHash` in the output in the line that starts with `got:` and copy it into this file.
        go-ethereum =
          let
            version = "1.13.4";
            src = pkgs.fetchFromGitHub {
              owner = "ethereum";
              repo = "go-ethereum";
              rev = "v${version}";
              hash = "sha256-RQlWWHoij3gtFwjJeEGsmd5YJNTGX0I84nOAQyWBx/M=";
            };
            vendorHash = "sha256-YmUgKO3JtVOE/YACqL/QBiyR1jT/jPCH+Gb0xYwkJEc=";
          in
          pkgs.go-ethereum.override {
            buildGoModule = args: pkgs.buildGoModule (args // {
              inherit version src vendorHash;
            });
          };
      in
      {
        devShells.default = pkgs.mkShell {
          COMPOSE_DOCKER_CLI_BUILD = 1;
          DOCKER_BUILDKIT = 1;
          packages = with pkgs; [
            nixWithFlakes
            go
            # goimports, godoc, etc.
            gotools
            # https://github.com/golangci/golangci-lint
            golangci-lint

            # Node
            nodejs
            pnpm
            yarn # `pnpm build` fails without this

            # Foundry, and tools like the anvil dev node
            foundry
            # solc

            # Docker
            docker-compose # provides the `docker-compose` command
            docker-buildx

            # Python
            (python3.withPackages (ps: with ps; [ ]))
            jq

            # geth node
            go-ethereum
          ] ++ lib.optionals stdenv.isDarwin [
            darwin.libobjc
            darwin.IOKit
            darwin.apple_sdk.frameworks.CoreFoundation
          ];
        };
      });
}
