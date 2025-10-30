package main

import (
	"strings"
	"testing"
)

func TestUnmarshalVarsConfigEmpty(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(""))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 0 {
		t.Errorf("expected empty config, but got: %v", *config)
	}
}

func TestUnmarshalVarsConfigSingleEntry(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  aaa: AAA
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 variable, but got: %v", *config)
	}
	fooVar, exists := (*config)["FOO_VAR"]
	if !exists {
		t.Fatalf("expected FOO_VAR to exist")
	}
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Path != "aaa" {
		t.Fatalf("expected Path %v, but got: %v", "aaa", fooVar[0].Path)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value %v, but got: %v", "AAA", *fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigContainingEmptyLines(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  aaa: AAA

  bbb: BBB
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 variable, but got: %v", *config)
	}
	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 2 {
		t.Fatalf("expected FOO_VAR to have 2 entries, but got: %v", fooVar)
	}
	if fooVar[0].Path != "aaa" {
		t.Fatalf("expected Path: aaa, but got: %v", fooVar[0].Path)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value: AAA, but got: %v", *fooVar[0].Value)
	}
	if fooVar[1].Path != "bbb" {
		t.Fatalf("expected Path: bbb, but got: %v", fooVar[1].Path)
	}
	if *fooVar[1].Value != "BBB" {
		t.Fatalf("expected Value=BBB, but got: %v", *fooVar[1].Value)
	}
}

func TestUnmarshalVarsConfigContainingComments(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
# This is a comment
FOO_VAR:
  aaa: AAA
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 variable, but got: %v", *config)
	}
	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Path != "aaa" {
		t.Fatalf("expected Path: aaa, but got: %v", fooVar[0].Path)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value: AAA, but got: %v", *fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigDoubleQuotedString(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
"FOO_VAR":
  "aaa/bbb": "AAA"
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 variable, but got: %v", *config)
	}
	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Path != "aaa/bbb" {
		t.Fatalf("expected Path: aaa/bbb, but got: %v", fooVar[0].Path)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value: AAA, but got: %v", *fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigMultipleVariables(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  path/to/dir: foo-value-1
  other/path: foo-value-2
BAR_VAR:
  another/dir: bar-value-1
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 2 {
		t.Fatalf("expected config with 2 variables, but got: %v", *config)
	}

	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 2 {
		t.Fatalf("expected FOO_VAR to have 2 entries, but got: %v", fooVar)
	}

	barVar := (*config)["BAR_VAR"]
	if len(barVar) != 1 {
		t.Fatalf("expected BAR_VAR to have 1 entry, but got: %v", barVar)
	}
}

func TestUnmarshalVarsConfigNullValue(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  path/to/dir: null
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Value != nil {
		t.Fatalf("expected Value to be nil, but got: %v", fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigEmptyValue(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  path/to/dir:
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	fooVar := (*config)["FOO_VAR"]
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Value != nil {
		t.Fatalf("expected Value to be nil, but got: %v", fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigValueContainingColon(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  aaa: "AAA:BBB"
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 variable, but got: %v", *config)
	}
	fooVar, exists := (*config)["FOO_VAR"]
	if !exists {
		t.Fatalf("expected FOO_VAR to exist")
	}
	if len(fooVar) != 1 {
		t.Fatalf("expected FOO_VAR to have 1 entry, but got: %v", fooVar)
	}
	if fooVar[0].Path != "aaa" {
		t.Fatalf("expected Path: %v, but got: %v", "aaa", fooVar[0].Path)
	}
	if *fooVar[0].Value != "AAA:BBB" {
		t.Fatalf("expected Value: %v, but got: %v", "AAA:BBB", *fooVar[0].Value)
	}
}

func TestUnmarshalVarsConfigExec(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
FOO_VAR:
  /tmp/example:
    echo: foo
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	fooVar, ok := (*config)["FOO_VAR"]
	if !ok {
		t.Fatalf("expected FOO_VAR to exist")
	}
	if len(fooVar) != 1 {
		t.Fatalf("expected one entry, got %d", len(fooVar))
	}
	if fooVar[0].Path != "/tmp/example" {
		t.Fatalf("unexpected Path: %s", fooVar[0].Path)
	}
	if fooVar[0].Value != nil {
		t.Fatalf("Value should be nil when exec is set")
	}
	if fooVar[0].Exec == nil {
		t.Fatalf("Exec should not be nil")
	}
	if fooVar[0].Exec.Id != "echo" {
		t.Fatalf("unexpected Exec.Id: %s", fooVar[0].Exec.Id)
	}
	if len(fooVar[0].Exec.Args) != 1 || fooVar[0].Exec.Args[0] != "foo" {
		t.Fatalf("unexpected Exec.Args: %#v", fooVar[0].Exec.Args)
	}
}

func TestUnmarshalVarsConfigExecMultipleArgs(t *testing.T) {
	config, err := UnmarshalVarsConfig([]byte(strings.TrimSpace(`
ECHO_VAR:
  some/dir:
    echo: [ John, Alice ]
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	echoVar, ok := (*config)["ECHO_VAR"]
	if !ok {
		t.Fatalf("expected ECHO_VAR to exist")
	}
	if len(echoVar) != 1 {
		t.Fatalf("expected one entry, got %d", len(echoVar))
	}
	if echoVar[0].Path != "some/dir" {
		t.Fatalf("unexpected Path: %s", echoVar[0].Path)
	}
	if echoVar[0].Value != nil {
		t.Fatalf("Value should be nil when exec reference is set")
	}
	if echoVar[0].Exec == nil {
		t.Fatalf("Exec should not be nil")
	}
	if echoVar[0].Exec.Id != "echo" {
		t.Fatalf("unexpected Exec.Id: %s", echoVar[0].Exec.Id)
	}
	if len(echoVar[0].Exec.Args) != 2 || echoVar[0].Exec.Args[0] != "John" || echoVar[0].Exec.Args[1] != "Alice" {
		t.Fatalf("unexpected Exec.Args: %#v", echoVar[0].Exec.Args)
	}
}

func TestUnmarshalExecsConfig(t *testing.T) {
	config, err := UnmarshalExecsConfig([]byte(strings.TrimSpace(`
gh: gh auth token --user %s
echo: bash -c 'echo %s and %s'
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(*config))
	}
	ghCmd, ok := (*config)["gh"]
	if !ok {
		t.Fatalf("expected 'gh' entry to exist")
	}
	if ghCmd != "gh auth token --user %s" {
		t.Fatalf("unexpected gh command: %s", ghCmd)
	}
	echoCmd, ok := (*config)["echo"]
	if !ok {
		t.Fatalf("expected 'echo' entry to exist")
	}
	if echoCmd != "bash -c 'echo %s and %s'" {
		t.Fatalf("unexpected echo command: %s", echoCmd)
	}
}

func TestUnmarshalExecsConfigEmpty(t *testing.T) {
	config, err := UnmarshalExecsConfig([]byte(""))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 0 {
		t.Errorf("expected empty config, but got: %v", *config)
	}
}
