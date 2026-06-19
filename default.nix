{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
    import (fetchTree nixpkgs.locked) {
      overlays = [
        (import "${fetchTree gomod2nix.locked}/overlay.nix")
      ];
    }
  ),
  buildGoApplication ? pkgs.buildGoApplication,
}:

buildGoApplication {
  pname = "chelly";
  version = "0.0.1";
  pwd = ./.;
  src = pkgs.lib.cleanSourceWith {
    src = ./.;
    filter =
      path: type:
      let
        baseName = baseNameOf path;
      in
      type == "regular"
      && (pkgs.lib.hasSuffix baseName ".go" || baseName == "go.mod" || baseName == "go.sum");
  };
  nativeBuildInputs = [ ];
  subPackages = [ "." ];
}
