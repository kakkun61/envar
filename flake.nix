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
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
      home-manager,
      ...
    }:
    let
      homeModule =
        {
          pkgs,
          lib,
          config,
          ...
        }:
        let
          config' = config.programs.envar;
          yamlFormat = pkgs.formats.yaml { };
        in
        {
          options.programs.envar = {
            enable = lib.mkEnableOption "Enable envar";
            package = lib.mkOption {
              type = with lib.types; package;
              default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
              description = "The envar package to use.";
            };
            enableBashIntegration = lib.mkEnableOption "Enable bash integration";
            settings = {
              vars = lib.mkOption {
                type =
                  with lib.types;
                  # var
                  attrsOf (
                    # path
                    attrsOf (
                      either
                        (
                          # value
                          str
                        )
                        (
                          # command
                          attrsOf (
                            either str (listOf str) # args
                          )
                        )
                    )
                  );
                default = { };
                description = "Environment variables to set.";
              };
              execs = lib.mkOption {
                type = with lib.types; attrsOf str;
                default = { };
                description = "Scripts to execute";
              };
            };
          };
          config = lib.mkIf config'.enable {
            home.packages = [ config'.package ];
            xdg.configFile = {
              "envar/vars.yaml".source = yamlFormat.generate "envar-vars.yaml" config'.settings.vars;
              "envar/execs.yaml".source = yamlFormat.generate "envar-execs.yaml" config'.settings.execs;
            };
            programs.bash = lib.mkIf config'.enableBashIntegration {
              initExtra = ''
                eval "$(${config'.package}/bin/envar hook)"
              '';
              logoutExtra = ''
                ${config'.package}/bin/envar hook logout $$
              '';
            };
          };
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        treefmt-nix.flakeModule
        home-manager.flakeModules.home-manager
      ];
      systems = nixpkgs.lib.systems.flakeExposed;
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
              cspell
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
        };
      flake = {
        homeModules.default = homeModule;
        homeConfigurations.test = home-manager.lib.homeManagerConfiguration {
          pkgs = import nixpkgs { system = "x86_64-linux"; };
          modules = [
            homeModule
            {
              home = {
                stateVersion = "24.05";
                username = "test";
                homeDirectory = "/home/test";
              };
              programs.envar = {
                enable = true;
                enableBashIntegration = true;
                settings = {
                  vars = {
                    FOO = {
                      "/tmp/foo" = "foo";
                    };
                  };
                  execs = {
                    bar = "/bin/bar";
                  };
                };
              };
            }
          ];
        };
      };
    };
}
