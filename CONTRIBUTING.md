# envar

## Build

There are multiple ways to build the project:

```
go build .
```

```
nix build
```

## Test

Testing Go code to run:

```
go test .
```

Testing a home-manager module to run:

```
home-manager build --flake .#test
```
