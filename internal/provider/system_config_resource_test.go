package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestSystemConfigResourceMetadata(t *testing.T) {
	r := NewSystemConfigResource()
	req := resource.MetadataRequest{ProviderTypeName: "goalert"}
	resp := resource.MetadataResponse{}
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "goalert_system_config" {
		t.Fatalf("expected type name goalert_system_config, got %s", resp.TypeName)
	}
}

func TestSystemConfigResourceSchema(t *testing.T) {
	r := NewSystemConfigResource()
	req := resource.SchemaRequest{}
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	attrs := resp.Schema.Attributes
	for _, expected := range []string{"id", "settings", "secrets_wo", "secret_hashes"} {
		if _, ok := attrs[expected]; !ok {
			t.Fatalf("missing required schema attribute: %s", expected)
		}
	}

	secretsWO, ok := attrs["secrets_wo"].(schema.MapAttribute)
	if !ok || !secretsWO.IsWriteOnly() {
		t.Fatalf("expected secrets_wo to be a WriteOnly MapAttribute")
	}
}

func TestSystemConfigKeyValidation(t *testing.T) {
	if len(systemConfigSettingsKeys) != 59 {
		t.Fatalf("expected 59 non-password settings keys, got %d", len(systemConfigSettingsKeys))
	}
	if len(systemConfigSecretKeys) != 9 {
		t.Fatalf("expected 9 password secret keys, got %d", len(systemConfigSecretKeys))
	}

	// Valid partition passes
	if err := validateSystemConfigMaps(
		map[string]string{"General.PublicURL": "https://alerts.example.com", "OIDC.Enable": "true"},
		map[string]string{"OIDC.ClientSecret": "secret-123"},
	); err != nil {
		t.Fatalf("expected valid maps to pass, got: %v", err)
	}

	// Secret key inside settings is rejected
	err := validateSystemConfigMaps(
		map[string]string{"OIDC.ClientSecret": "secret-123"},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "secrets_wo") {
		t.Fatalf("expected error directing OIDC.ClientSecret to secrets_wo, got: %v", err)
	}

	// Non-secret key inside secrets_wo is rejected
	err = validateSystemConfigMaps(
		nil,
		map[string]string{"General.PublicURL": "https://alerts.example.com"},
	)
	if err == nil || !strings.Contains(err.Error(), "settings") {
		t.Fatalf("expected error directing General.PublicURL to settings, got: %v", err)
	}

	// Unknown key is rejected
	err = validateSystemConfigMaps(
		map[string]string{"NonExistent.Key": "val"},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "NonExistent.Key") {
		t.Fatalf("expected unknown key error, got: %v", err)
	}
}

func TestParseSystemConfigImportID(t *testing.T) {
	settingsKeys, secretKeys, err := parseSystemConfigImportID("system")
	if err != nil || len(settingsKeys) != 0 || len(secretKeys) != 0 {
		t.Fatalf("unexpected bare system import result: settings=%v secrets=%v err=%v", settingsKeys, secretKeys, err)
	}

	settingsKeys, secretKeys, err = parseSystemConfigImportID("system:OIDC.Enable,OIDC.IssuerURL,OIDC.ClientSecret")
	if err != nil {
		t.Fatalf("unexpected compound import error: %v", err)
	}
	if len(settingsKeys) != 2 || settingsKeys[0] != "OIDC.Enable" || settingsKeys[1] != "OIDC.IssuerURL" {
		t.Fatalf("unexpected settingsKeys: %v", settingsKeys)
	}
	if len(secretKeys) != 1 || secretKeys[0] != "OIDC.ClientSecret" {
		t.Fatalf("unexpected secretKeys: %v", secretKeys)
	}

	if _, _, err := parseSystemConfigImportID("invalid"); err == nil {
		t.Fatalf("expected error on invalid import ID")
	}
	if _, _, err := parseSystemConfigImportID("system:Unknown.Key"); err == nil {
		t.Fatalf("expected error on unknown key in import ID")
	}
}
