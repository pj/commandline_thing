{
  description = "Flake command line thing development";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
        commandline_thing = pkgs.buildGoApplication {
          pname = "commandline_thing";
          version = "1.0.0";
          src = ./pkg;
          modules = ./gomod2nix.toml;
        };
      in
      {
        packages = {
          commandline_thing = commandline_thing;
        };
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            delve
            gopls
            go-tools
            gomod2nix.packages.${system}.default
          ];
        };
      }
    );
}
