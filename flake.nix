{ description = "Railpack";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { flake-utils, nixpkgs, ... }:
    flake-utils.lib.eachDefaultSystem (arch: let
      pkgs = import nixpkgs { system = arch; };
    in {
      packages.default = pkgs.buildGoModule {
        pname = "railpack";
        version = "0.9.0";
        src = ./.;
        subPackages = [ "cmd/cli" ];
        vendorHash = "sha256-bn6GsJBRg4S5IWBShlYXk12nNuAnv4MmKZvxE0sujT8=";
      };
  });
}
