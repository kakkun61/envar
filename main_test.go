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
	config, err := UnmarshalConfig([]byte("aaa: AAA"))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 entry, but got: %v", *config)
	}
	expected := ConfigPair{PathPrefix: "aaa", User: "AAA"}
	if (*config)[0] != expected {
		t.Fatalf("expected config entry %v, but got: %v", expected, (*config)[0])
	}
}

func TestUnmarshalConfigContainingEmptyLines(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
aaa: AAA

bbb: BBB
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 2 {
		t.Fatalf("expected config with 2 entries, but got: %v", *config)
	}
	expected1 := ConfigPair{PathPrefix: "aaa", User: "AAA"}
	if (*config)[0] != expected1 {
		t.Fatalf("expected config entry %v, but got: %v", expected1, (*config)[0])
	}
	expected2 := ConfigPair{PathPrefix: "bbb", User: "BBB"}
	if (*config)[1] != expected2 {
		t.Fatalf("expected config entry %v, but got: %v", expected2, (*config)[1])
	}
}

func TestUnmarshalConfigContainingComments(t *testing.T) {
	config, err := UnmarshalConfig([]byte(strings.TrimSpace(`
# This is a comment
aaa: AAA
	`)))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 entry, but got: %v", *config)
	}
	expected := ConfigPair{PathPrefix: "aaa", User: "AAA"}
	if (*config)[0] != expected {
		t.Fatalf("expected config entry %v, but got: %v", expected, (*config)[0])
	}
}

func TestUnmarshalConfigDoubleQuotedString(t *testing.T) {
	config, err := UnmarshalConfig([]byte(`"aaa/bbb": AAA`))
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if len(*config) != 1 {
		t.Fatalf("expected config with 1 entry, but got: %v", *config)
	}
	expected := ConfigPair{PathPrefix: "aaa/bbb", User: "AAA"}
	if (*config)[0] != expected {
		t.Fatalf("expected config entry %v, but got: %v", expected, (*config)[0])
	}
}
