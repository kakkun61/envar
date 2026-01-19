{
  description = "envar";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    home-manager.url = "github:nix-community/home-manager";
    cspell-nix = {
      url = "github:kakkun61/cspell-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
      home-manager,
      cspell-nix,
      ...
    }:
    let
      homeModule = nixpkgs.lib.setDefaultModuleLocation ./home-module.nix (
        import ./home-module.nix { inherit inputs; }
      );
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        flake-parts.flakeModules.modules
        treefmt-nix.flakeModule
        home-manager.flakeModules.home-manager
        cspell-nix.flakeModule
      ];
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];
      perSystem =
        {
          config,
          self',
          inputs',
          pkgs,
          system,
          ...
        }:
        {
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
            ];
          };
          packages.default = pkgs.buildGoModule {
            pname = "envar";
            version = "1";
            src = ./.;
            vendorHash = "sha256-eoJHYEEZYbD/IYar7JhbyuWWjSo7fkJoNNnVDwOVeV4=";
          };
          treefmt = {
            programs = {
              nixfmt.enable = true;
              yamlfmt.enable = true;
            };
          };
          cspell.configFile = ./cspell.yaml;
        };
      flake = {
        homeModules.default = homeModule;
        homeConfigurations.test = import ./home-configuration-test.nix { inherit inputs homeModule; };
        modules.homeManager.default = homeModule;
      };
    };
}
