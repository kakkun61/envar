# ./home-configuration-test.nix returns a Home Manager configuration
{ inputs, system, homeModule }:
let
  inherit (inputs) nixpkgs home-manager;
in
home-manager.lib.homeManagerConfiguration {
  pkgs = import nixpkgs { inherit system; };
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
            FOO = [
              {
                path = "/tmp/foo";
                value = "foo";
              }
              {
                path = "/tmp/bar";
                value = {
                  gh = "kakkun61";
                };
              }
            ];
          };
          execs = {
            bar = "/bin/bar";
          };
        };
      };
    }
  ];
}
