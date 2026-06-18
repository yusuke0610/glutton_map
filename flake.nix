{
  description = "glutton_map";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go gopls gotools        # backend (go 1.22+)
            bun nodejs_22           # web + openapi-typescript
            docker-compose
          ];
          shellHook = ''
            echo "如月愛マップ: backend=go / web=bun"
          '';
        };
      });
}
