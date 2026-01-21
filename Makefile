doc/options.md: dev/make-options-doc flake.nix home-module.nix
	-mkdir -p $(@D)
	./dev/make-options-doc > $@

.PHONY: home-module.test
home-module.test:
	home-manager build --flake .#test-$$(nix eval --impure --expr 'builtins.currentSystem')
