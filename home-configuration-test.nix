# ./home-configuration-test.nix returns a Home Manager configuration
{ inputs, homeModule }:
let
  inherit (inputs) nixpkgs home-manager;
in
home-manager.lib.homeManagerConfiguration {
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
}
