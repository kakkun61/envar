# envar

This is a command-line tool that automatically switches values of environment variables based on the current directory path.

1. Place the `envar` binary
2. Install the shell hook
3. Write the configuration file

## Place the `envar` binary

Run `go build` and place the binary file in your `PATH`.

## Install the shell hook

Add this line to your _.bashrc_:

```bash
eval "$(envar hook)"
```

and add this line to your _.bash_logout_:

```bash
envar hook logout $$
```

## Write the configuration file

The configuration file uses a subset of the YAML format. It is located at _`$CONFIG_DIR`/envar/config.yaml_. `$CONFIG_DIR` is a returned value of [`os.UserConfigDir()`](https://pkg.go.dev/os#UserConfigDir).

The file format consists of two levels. The first level is an environment variable name and colon. The second level is a list of lines containing two fields: a path prefix and a GitHub username separated by a colon. Comment lines start with `#` and empty lines are ignored. Path prefixes can be quoted with double quotes. `~` expands to the user's home directory. Lines are matched from top to bottom, with the first match taking precedence.

For instance:

```yaml
FOO_VAR:
  path/to/dir: foo-value-1
  other/path: foo-value-2
  other/path/never: foo-value-3
  # This is a comment line
  ~/projects: foo-value-4
  "spacial dir/path": foo-value-5
BAR_VAR:
  another/dir: bar-value-1
```

The directory _path/to/dir_ and its subdirectories are associated with `foo-value-1`. The directory _other/path_ and its subdirectories are associated with `foo-value-2`. `foo-value-3` is never used because _other/path/never_ is a subdirectory of _other/path_. The fourth line is a comment and is ignored. The fifth line demonstrates the use of `~`. The sixth line shows an example of using double quotes.

You can unset a variable by specifying a `null` value:

```yaml
FOO_VAR:
  path/to/dir:
# or
  path/to/dir: null
```

When no matching path prefix is found for a variable, it is unset.

You can use programs to get values, if you don't want to write some secret values directly in the configuration file for example. Here is an instance using `gh` CLI to get GitHub authentication token:

```yaml
GH_TOKEN:
  path/to/dir:
    exec: gh auth token --user foo
  other/path:
    exec: gh auth token --user bar
```
