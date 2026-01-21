# `pkgs.callPackage ./options-doc.nix { inherit inputs homeModule; }` returns a
# derivation that produces a module options documentation.
{
  inputs,
  homeModule,
  pkgs,
  lib,
}:
let
  eval = lib.evalModules {
    modules = [
      { _module.args = { inherit pkgs; }; }
      (
        arg@{
          lib,
          config,
          pkgs,
          ...
        }:
        {
          inherit
            (homeModule {
              inherit config pkgs;
              lib = lib // inputs.home-manager.lib;
            })
            options
            ;
        }
      )
    ];
    class = "homeManager";
  };
  doc = pkgs.nixosOptionsDoc { inherit (eval) options; };
in
doc.optionsCommonMark
