package main

import (
	"strings"
	"testing"
)

func TestUnmarshalConfigEmpty(t *testing.T) {
	config, err := UnmarshalConfig([]byte(""))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 0 {
		t.Errorf("expected empty config, but got: %v", *config)
	}
}

func TestUnmarshalConfigSingleEntry(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
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
	if fooVar[0].PathPrefix != "aaa" {
		t.Fatalf("expected PathPrefix %v, but got: %v", "aaa", fooVar[0].PathPrefix)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value %v, but got: %v", "AAA", *fooVar[0].Value)
	}
}

func TestUnmarshalConfigContainingEmptyLines(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
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
	if fooVar[0].PathPrefix != "aaa" {
		t.Fatalf("expected PathPrefix=aaa, but got: %v", fooVar[0].PathPrefix)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value=AAA, but got: %v", *fooVar[0].Value)
	}
	if fooVar[1].PathPrefix != "bbb" {
		t.Fatalf("expected PathPrefix=bbb, but got: %v", fooVar[1].PathPrefix)
	}
	if *fooVar[1].Value != "BBB" {
		t.Fatalf("expected Value=BBB, but got: %v", *fooVar[1].Value)
	}
}

func TestUnmarshalConfigContainingComments(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
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
	if fooVar[0].PathPrefix != "aaa" {
		t.Fatalf("expected PathPrefix=aaa, but got: %v", fooVar[0].PathPrefix)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value=AAA, but got: %v", *fooVar[0].Value)
	}
}

func TestUnmarshalConfigDoubleQuotedString(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
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
	if fooVar[0].PathPrefix != "aaa/bbb" {
		t.Fatalf("expected PathPrefix=aaa/bbb, but got: %v", fooVar[0].PathPrefix)
	}
	if *fooVar[0].Value != "AAA" {
		t.Fatalf("expected Value=AAA, but got: %v", *fooVar[0].Value)
	}
}

func TestUnmarshalConfigMultipleVariables(t *testing.T) {
	config, err := UnmarshalConfig([]byte(`FOO_VAR:
  path/to/dir: foo-value-1
  other/path: foo-value-2
BAR_VAR:
  another/dir: bar-value-1`))
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

func TestUnmarshalConfigNullValue(t *testing.T) {
	config, err := UnmarshalConfig([]byte(`FOO_VAR:
  path/to/dir: null`))
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

func TestUnmarshalConfigEmptyValue(t *testing.T) {
	config, err := UnmarshalConfig([]byte(`FOO_VAR:
  path/to/dir:`))
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

func TestUnmarshalConfigValueContainingColon(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
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
	if fooVar[0].PathPrefix != "aaa" {
		t.Fatalf("expected PathPrefix %v, but got: %v", "aaa", fooVar[0].PathPrefix)
	}
	if *fooVar[0].Value != "AAA:BBB" {
		t.Fatalf("expected Value %v, but got: %v", "AAA:BBB", *fooVar[0].Value)
	}
}
