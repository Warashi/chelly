{
  inputs = {
    # keep-sorted start block=yes
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };
    # keep-sorted end
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        inputs.flake-parts.flakeModules.partitions
      ];

      partitionedAttrs = {
        # keep-sorted start
        checks = "dev";
        devShells = "dev";
        formatter = "dev";
        # keep-sorted end
      };
      partitions.dev = {
        extraInputsFlake = ./nix/dev;
        module = ./nix/dev/flake-module.nix;
      };

      perSystem =
        {
          inputs',
          system,
          pkgs,
          ...
        }:
        {
          packages = rec {
            default = chelly;
            chelly = pkgs.callPackage ./. {
              inherit (inputs'.gomod2nix.legacyPackages) buildGoApplication;
            };
          };
        };

      systems = [
        "aarch64-linux"
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
    };
}
