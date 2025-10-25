# gh-auth-switch

This is a command-line tool that automatically switches GitHub CLI authentication users based on the current directory path.

1. Place the `gh-auth-switch` binary
2. Install the shell hook
3. Write the configuration file

## Place the `gh-auth-switch` binary

Run `go build` and place the binary file in your PATH.

## Install the shell hook

Add this line to your _.bashrc_:

```bash
eval "$(gh-auth-switch hook)"
```

## Write the configuration file

The configuration file uses a subset of the YAML format. It is located at _`$CONFIG_DIR`/gh-auth-switch/config.yaml_. `$CONFIG_DIR` is 

The file format consists of lines containing two fields: a path prefix and a GitHub username separated by a colon. Comment lines start with `#` and empty lines are ignored. Path prefixes can be quoted with double quotes. `~` expands to the user's home directory. Lines are matched from top to bottom, with the first match taking precedence.

For instance:

```yaml
path/to/dir: github-user1
other/path: github-user2
other/path/never: github-user3
# This is a comment line
~/projects: github-user4
"spacial dir/path": github-user5
```

The directory _path/to/dir_ and its subdirectories are associated with `github-user1`. The directory _other/path_ and its subdirectories are associated with `github-user2`. `github-user3` is never used because _other/path/never_ is a subdirectory of _other/path_. The fourth line is a comment and is ignored. The fifth line demonstrates the use of `~`. The sixth line shows an example of using double quotes.
