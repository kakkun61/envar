.PHONY: home-module.test
home-module.test:
	home-manager build --flake .#test-$$(nix eval --impure --expr 'builtins.currentSystem')
