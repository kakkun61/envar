# home-module.nix returns a Home Manager module
{ inputs }:
let
  inherit (inputs) self;
in
{
  pkgs,
  lib,
  config,
  ...
}:
let
  config' = config.programs.envar;
  yamlFormat = pkgs.formats.yaml { };
  inherit (import ./internal.nix { inherit lib; }) makeVarsYamlString;
in
{
  _class = "homeManager";

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
            listOf (submodule {
              options = {
                path = lib.mkOption {
                  type = str;
                  description = "Path pattern to match.";
                };
                value = lib.mkOption {
                  type =
                    either
                      (
                        # value
                        str
                      )
                      (
                        # command
                        attrsOf (
                          # args
                          either str (listOf str)
                        )
                      );
                  description = "Value or command to bind the variable to.";
                };
              };
            })
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
      "envar/vars.yaml".text = makeVarsYamlString config'.settings.vars;
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
}
