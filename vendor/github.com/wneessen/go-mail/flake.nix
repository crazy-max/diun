{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    go-overlay.url = "github:purpleclay/go-overlay";
  };

  outputs = { nixpkgs, flake-utils, go-overlay, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ go-overlay.overlays.default ];
        };
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            go-overlay.packages.${system}.govendor
            nixpkgs.legacyPackages.${system}.dkimpy
            nixpkgs.legacyPackages.${system}.gofumpt
            nixpkgs.legacyPackages.${system}.golangci-lint
            nixpkgs.legacyPackages.${system}.reuse
          ];
        };
      }
    );
}
