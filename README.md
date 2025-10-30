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

The configuration file uses YAML. It is located at _`$CONFIG_DIR`/envar/vars.yaml_ and _`$CONFIG_DIR`/envar/execs.yaml_. `$CONFIG_DIR` is the value returned by [`os.UserConfigDir()`](https://pkg.go.dev/os#UserConfigDir).

_vars.yaml_ is used to define environment variable values. For example:

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

You can compute values using a command, which is useful when you don't want to store secrets directly in the configuration file. For example, using the `gh` CLI to get a GitHub authentication token, you must prepare _execs.yaml_ first like this:

```yaml
gh: gh auth token --user %s
```

The key `gh` is a config variable that is referenced in _vars.yaml_ later, and the value is a command template. The placeholder `%s` is replaced by the argument specified in _vars.yaml_.

Then, you can use it in _vars.yaml_ like this:

```yaml
GH_TOKEN:
  path/to/dir:
    gh: foo
  other/path:
    gh: bar
```

Placeholders can be placed multiple times in a command template, for example when _execs.yaml_ is like this:

```yaml
echo: bash -c 'echo %s and %s'
```

You can use it in _vars.yaml_ like this:

```yaml
ECHO_VAR:
  some/dir:
    echo: [ John, Alice ]
```

Note that no escaping is performed for the arguments.
