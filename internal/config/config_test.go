package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func TestLoad_ParsesAllSections(t *testing.T) {
	dir := writeConfig(t, `
params:
  TOKEN:
    description: a token
    required: true
  REGION:
    default: eu-west
copy:
  - .env
pre_create:
  - git pull
post_create:
  - docker compose build
pre_remove:
  - docker compose down
`)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Params["TOKEN"].Required {
		t.Error("TOKEN should be required")
	}
	if cfg.Params["REGION"].Default != "eu-west" {
		t.Errorf("REGION default = %q", cfg.Params["REGION"].Default)
	}
	if !reflect.DeepEqual(cfg.Copy, []string{".env"}) {
		t.Errorf("Copy = %v", cfg.Copy)
	}
	if len(cfg.PreCreate) != 1 || cfg.PreCreate[0] != "git pull" {
		t.Errorf("PreCreate = %v", cfg.PreCreate)
	}
	if len(cfg.PostCreate) != 1 {
		t.Errorf("PostCreate = %v", cfg.PostCreate)
	}
	if len(cfg.PreRemove) != 1 {
		t.Errorf("PreRemove = %v", cfg.PreRemove)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := writeConfig(t, "params: [this is not a map")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestResolveParams_AppliesDefaults(t *testing.T) {
	cfg := &Config{Params: map[string]Param{
		"REGION": {Default: "eu-west"},
	}}
	got, err := cfg.ResolveParams(nil)
	if err != nil {
		t.Fatalf("ResolveParams: %v", err)
	}
	if got["REGION"] != "eu-west" {
		t.Errorf("REGION = %q", got["REGION"])
	}
}

func TestResolveParams_UserOverridesDefault(t *testing.T) {
	cfg := &Config{Params: map[string]Param{
		"REGION": {Default: "eu-west"},
	}}
	got, err := cfg.ResolveParams(map[string]string{"REGION": "us-east"})
	if err != nil {
		t.Fatalf("ResolveParams: %v", err)
	}
	if got["REGION"] != "us-east" {
		t.Errorf("REGION = %q, want us-east", got["REGION"])
	}
}

func TestResolveParams_MissingRequired(t *testing.T) {
	cfg := &Config{Params: map[string]Param{
		"TOKEN": {Required: true},
	}}
	if _, err := cfg.ResolveParams(nil); err == nil {
		t.Fatal("expected error for missing required param")
	}
}

func TestResolveParams_RequiredSatisfied(t *testing.T) {
	cfg := &Config{Params: map[string]Param{
		"TOKEN": {Required: true},
	}}
	got, err := cfg.ResolveParams(map[string]string{"TOKEN": "abc"})
	if err != nil {
		t.Fatalf("ResolveParams: %v", err)
	}
	if got["TOKEN"] != "abc" {
		t.Errorf("TOKEN = %q", got["TOKEN"])
	}
}
