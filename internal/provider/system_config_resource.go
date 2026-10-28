package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/healdropper/terraform-provider-goalert/internal/client"
)

const systemConfigPrivateSaltKey = "system_config_salt"

var systemConfigSecretKeys = map[string]struct{}{
	"GitHub.ClientSecret":       {},
	"OIDC.ClientSecret":         {},
	"Mailgun.APIKey":            {},
	"Slack.ClientSecret":        {},
	"Slack.AccessToken":         {},
	"Slack.SigningSecret":       {},
	"Twilio.AuthToken":          {},
	"Twilio.AlternateAuthToken": {},
	"SMTP.Password":             {},
}

var systemConfigSettingsKeys = map[string]struct{}{
	"General.ApplicationName":              {},
	"General.PublicURL":                    {},
	"General.GoogleAnalyticsID":            {},
	"General.NotificationDisclaimer":       {},
	"General.DisableMessageBundles":        {},
	"General.ShortURL":                     {},
	"General.DisableSMSLinks":              {},
	"General.DisableLabelCreation":         {},
	"General.DisableCalendarSubscriptions": {},
	"Services.RequiredLabels":              {},
	"Maintenance.AlertCleanupDays":         {},
	"Maintenance.AlertAutoCloseDays":       {},
	"Maintenance.AutoCloseAckedAlerts":     {},
	"Maintenance.APIKeyExpireDays":         {},
	"Maintenance.ScheduleCleanupDays":      {},
	"Auth.RefererURLs":                     {},
	"Auth.DisableBasic":                    {},
	"GitHub.Enable":                        {},
	"GitHub.NewUsers":                      {},
	"GitHub.ClientID":                      {},
	"GitHub.AllowedUsers":                  {},
	"GitHub.AllowedOrgs":                   {},
	"GitHub.EnterpriseURL":                 {},
	"OIDC.Enable":                          {},
	"OIDC.NewUsers":                        {},
	"OIDC.OverrideName":                    {},
	"OIDC.IssuerURL":                       {},
	"OIDC.ClientID":                        {},
	"OIDC.Scopes":                          {},
	"OIDC.UserInfoEmailPath":               {},
	"OIDC.UserInfoEmailVerifiedPath":       {},
	"OIDC.UserInfoNamePath":                {},
	"Mailgun.Enable":                       {},
	"Mailgun.EmailDomain":                  {},
	"Slack.Enable":                         {},
	"Slack.ClientID":                       {},
	"Slack.InteractiveMessages":            {},
	"Slack.DisableBroadcastThreadReplies":  {},
	"Twilio.Enable":                        {},
	"Twilio.VoiceName":                     {},
	"Twilio.VoiceLanguage":                 {},
	"Twilio.AccountSID":                    {},
	"Twilio.FromNumber":                    {},
	"Twilio.MessagingServiceSID":           {},
	"Twilio.RCSSenderID":                   {},
	"Twilio.DisableTwoWaySMS":              {},
	"Twilio.SMSCarrierLookup":              {},
	"Twilio.SMSFromNumberOverride":         {},
	"SMTP.Enable":                          {},
	"SMTP.From":                            {},
	"SMTP.Address":                         {},
	"SMTP.DisableTLS":                      {},
	"SMTP.SkipVerify":                      {},
	"SMTP.Username":                        {},
	"Webhook.Enable":                       {},
	"Webhook.AllowedURLs":                  {},
	"Webhook.BlockPrivateAddresses":        {},
	"Feedback.Enable":                      {},
	"Feedback.OverrideURL":                 {},
}

type systemConfigResource struct {
	client *client.Client
}

type systemConfigModel struct {
	ID           types.String `tfsdk:"id"`
	Settings     types.Map    `tfsdk:"settings"`
	SecretsWO    types.Map    `tfsdk:"secrets_wo"`
	SecretHashes types.Map    `tfsdk:"secret_hashes"`
}

var (
	_ resource.Resource                = &systemConfigResource{}
	_ resource.ResourceWithConfigure   = &systemConfigResource{}
	_ resource.ResourceWithModifyPlan  = &systemConfigResource{}
	_ resource.ResourceWithImportState = &systemConfigResource{}
)

func NewSystemConfigResource() resource.Resource {
	return &systemConfigResource{}
}

func (r *systemConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_config"
}

func (r *systemConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages global system configuration settings and write-only secrets in GoAlert (Admin > Config).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Singleton identifier for the GoAlert system configuration ('system').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Non-secret GoAlert configuration key/value pairs (password: false). Removing a key unmanages it in Terraform; setting a key to \"\" clears it in GoAlert.",
			},
			"secrets_wo": schema.MapAttribute{
				Optional:            true,
				WriteOnly:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Write-only secret GoAlert configuration key/value pairs (password: true). Never persisted in Terraform state or plan artifacts.",
			},
			"secret_hashes": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Salted HMAC-SHA256 digests of managed secret keys in secrets_wo for automatic drift detection.",
			},
		},
	}
}

func (r *systemConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider client", "Expected a GoAlert API client.")
		return
	}
	r.client = c
}

func validateSystemConfigMaps(settings, secretsWO map[string]string) error {
	for k := range settings {
		if _, isSecret := systemConfigSecretKeys[k]; isSecret {
			return fmt.Errorf("config key %q is a secret (password: true) and must be configured in secrets_wo, not settings", k)
		}
		if _, ok := systemConfigSettingsKeys[k]; !ok {
			return fmt.Errorf("unknown GoAlert config key %q in settings", k)
		}
	}
	for k := range secretsWO {
		if _, isSetting := systemConfigSettingsKeys[k]; isSetting {
			return fmt.Errorf("config key %q is a non-secret setting (password: false) and must be configured in settings, not secrets_wo", k)
		}
		if _, ok := systemConfigSecretKeys[k]; !ok {
			return fmt.Errorf("unknown GoAlert secret config key %q in secrets_wo", k)
		}
	}
	return nil
}

func generateSystemConfigSalt() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (r *systemConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var config systemConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan systemConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue("system")

	var settingsMap map[string]string
	if !config.Settings.IsNull() && !config.Settings.IsUnknown() {
		settingsMap = make(map[string]string)
		for k, v := range config.Settings.Elements() {
			if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				settingsMap[k] = sv.ValueString()
			} else {
				settingsMap[k] = ""
			}
		}
	}

	var secretsMap map[string]string
	secretsKnown := true
	if !config.SecretsWO.IsNull() {
		if config.SecretsWO.IsUnknown() {
			secretsKnown = false
		} else {
			secretsMap = make(map[string]string)
			for k, v := range config.SecretsWO.Elements() {
				sv, ok := v.(types.String)
				if !ok || sv.IsUnknown() {
					secretsKnown = false
				} else if !sv.IsNull() {
					secretsMap[k] = sv.ValueString()
				}
			}
		}
	}

	if err := validateSystemConfigMaps(settingsMap, secretsMap); err != nil {
		resp.Diagnostics.AddError("Invalid system configuration keys", err.Error())
		return
	}

	if config.SecretsWO.IsNull() {
		plan.SecretHashes = types.MapNull(types.StringType)
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	if !secretsKnown {
		plan.SecretHashes = types.MapUnknown(types.StringType)
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		return
	}

	rawSalt, diags := req.Private.GetKey(ctx, systemConfigPrivateSaltKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	salt := string(rawSalt)
	if salt == "" {
		var err error
		salt, err = generateSystemConfigSalt()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate secret hash salt", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, systemConfigPrivateSaltKey, []byte(salt))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	hashElems := make(map[string]attr.Value, len(secretsMap))
	for k, v := range secretsMap {
		hashElems[k] = types.StringValue(client.HashSystemConfigSecret(salt, v))
	}
	hashesVal, mapDiags := types.MapValue(types.StringType, hashElems)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.SecretHashes = hashesVal
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func (r *systemConfigResource) applyConfig(ctx context.Context, plan systemConfigModel, config systemConfigModel, salt string) (systemConfigModel, error) {
	var settingsMap map[string]string
	if !plan.Settings.IsNull() && !plan.Settings.IsUnknown() {
		settingsMap = make(map[string]string, len(plan.Settings.Elements()))
		for k, v := range plan.Settings.Elements() {
			if sv, ok := v.(types.String); ok && !sv.IsNull() {
				settingsMap[k] = sv.ValueString()
			}
		}
	}

	var secretsMap map[string]string
	if !config.SecretsWO.IsNull() && !config.SecretsWO.IsUnknown() {
		secretsMap = make(map[string]string, len(config.SecretsWO.Elements()))
		for k, v := range config.SecretsWO.Elements() {
			if sv, ok := v.(types.String); ok && !sv.IsNull() {
				secretsMap[k] = sv.ValueString()
			}
		}
	}

	if err := validateSystemConfigMaps(settingsMap, secretsMap); err != nil {
		return plan, err
	}

	keys := make([]string, 0, len(settingsMap)+len(secretsMap))
	combined := make(map[string]string, len(settingsMap)+len(secretsMap))
	for k, v := range settingsMap {
		keys = append(keys, k)
		combined[k] = v
	}
	for k, v := range secretsMap {
		keys = append(keys, k)
		combined[k] = v
	}
	sort.Strings(keys)

	if len(keys) > 0 {
		inputs := make([]client.ConfigValueInput, 0, len(keys))
		for _, k := range keys {
			inputs = append(inputs, client.ConfigValueInput{ID: k, Value: combined[k]})
		}
		if err := r.client.SetSystemConfig(ctx, inputs); err != nil {
			return plan, err
		}
	}

	remoteItems, err := r.client.ReadSystemConfig(ctx, salt)
	if err != nil {
		return plan, err
	}
	remoteByID := make(map[string]client.SystemConfigValue, len(remoteItems))
	for _, item := range remoteItems {
		remoteByID[item.ID] = item
	}

	plan.ID = types.StringValue("system")
	plan.SecretsWO = types.MapNull(types.StringType)

	if plan.Settings.IsNull() {
		plan.Settings = types.MapNull(types.StringType)
	} else {
		elems := make(map[string]attr.Value, len(settingsMap))
		for k := range settingsMap {
			if item, ok := remoteByID[k]; ok {
				elems[k] = types.StringValue(item.Value)
			}
		}
		val, _ := types.MapValue(types.StringType, elems)
		plan.Settings = val
	}

	if config.SecretsWO.IsNull() {
		plan.SecretHashes = types.MapNull(types.StringType)
	} else {
		hashElems := make(map[string]attr.Value, len(secretsMap))
		for k := range secretsMap {
			if item, ok := remoteByID[k]; ok {
				hashElems[k] = types.StringValue(item.ValueHash)
			}
		}
		val, _ := types.MapValue(types.StringType, hashElems)
		plan.SecretHashes = val
	}

	return plan, nil
}

func (r *systemConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan systemConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config systemConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rawSalt, diags := resp.Private.GetKey(ctx, systemConfigPrivateSaltKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	salt := string(rawSalt)
	if salt == "" {
		var err error
		salt, err = generateSystemConfigSalt()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate secret hash salt", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, systemConfigPrivateSaltKey, []byte(salt))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	newState, err := r.applyConfig(ctx, plan, config, salt)
	if err != nil {
		resp.Diagnostics.AddError("Failed to apply GoAlert system configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *systemConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state systemConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rawSalt, diags := req.Private.GetKey(ctx, systemConfigPrivateSaltKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	salt := string(rawSalt)
	if salt == "" {
		var err error
		salt, err = generateSystemConfigSalt()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate secret hash salt", err.Error())
			return
		}
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, systemConfigPrivateSaltKey, []byte(salt))...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	remoteItems, err := r.client.ReadSystemConfig(ctx, salt)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read GoAlert system configuration", err.Error())
		return
	}
	remoteByID := make(map[string]client.SystemConfigValue, len(remoteItems))
	for _, item := range remoteItems {
		remoteByID[item.ID] = item
	}

	state.ID = types.StringValue("system")
	state.SecretsWO = types.MapNull(types.StringType)

	if !state.Settings.IsNull() && !state.Settings.IsUnknown() {
		elems := make(map[string]attr.Value, len(state.Settings.Elements()))
		for k := range state.Settings.Elements() {
			if item, ok := remoteByID[k]; ok {
				elems[k] = types.StringValue(item.Value)
			}
		}
		val, mapDiags := types.MapValue(types.StringType, elems)
		resp.Diagnostics.Append(mapDiags...)
		state.Settings = val
	}

	if !state.SecretHashes.IsNull() && !state.SecretHashes.IsUnknown() {
		hashElems := make(map[string]attr.Value, len(state.SecretHashes.Elements()))
		for k := range state.SecretHashes.Elements() {
			if item, ok := remoteByID[k]; ok {
				hashElems[k] = types.StringValue(item.ValueHash)
			}
		}
		val, mapDiags := types.MapValue(types.StringType, hashElems)
		resp.Diagnostics.Append(mapDiags...)
		state.SecretHashes = val
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *systemConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan systemConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var config systemConfigModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rawSalt, diags := req.Private.GetKey(ctx, systemConfigPrivateSaltKey)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	salt := string(rawSalt)
	if salt == "" {
		var err error
		salt, err = generateSystemConfigSalt()
		if err != nil {
			resp.Diagnostics.AddError("Failed to generate secret hash salt", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, systemConfigPrivateSaltKey, []byte(salt))...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.applyConfig(ctx, plan, config, salt)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update GoAlert system configuration", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *systemConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func parseSystemConfigImportID(rawID string) ([]string, []string, error) {
	if rawID == "system" {
		return nil, nil, nil
	}
	if !strings.HasPrefix(rawID, "system:") {
		return nil, nil, fmt.Errorf("expected import ID 'system' or 'system:<key1>,<key2>,...', got %q", rawID)
	}
	keyList := strings.TrimPrefix(rawID, "system:")
	if strings.TrimSpace(keyList) == "" {
		return nil, nil, fmt.Errorf("import ID %q has an empty key list after 'system:'", rawID)
	}
	var settingsKeys, secretKeys []string
	for _, rawKey := range strings.Split(keyList, ",") {
		k := strings.TrimSpace(rawKey)
		if k == "" {
			continue
		}
		if _, ok := systemConfigSettingsKeys[k]; ok {
			settingsKeys = append(settingsKeys, k)
		} else if _, ok := systemConfigSecretKeys[k]; ok {
			secretKeys = append(secretKeys, k)
		} else {
			return nil, nil, fmt.Errorf("unknown GoAlert config key %q in import ID", k)
		}
	}
	return settingsKeys, secretKeys, nil
}

func (r *systemConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settingsKeys, secretKeys, err := parseSystemConfigImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID for goalert_system_config", err.Error())
		return
	}

	salt, err := generateSystemConfigSalt()
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate secret hash salt", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, systemConfigPrivateSaltKey, []byte(salt))...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := systemConfigModel{
		ID:           types.StringValue("system"),
		Settings:     types.MapNull(types.StringType),
		SecretsWO:    types.MapNull(types.StringType),
		SecretHashes: types.MapNull(types.StringType),
	}

	if len(settingsKeys) > 0 {
		elems := make(map[string]attr.Value, len(settingsKeys))
		for _, k := range settingsKeys {
			elems[k] = types.StringValue("")
		}
		val, diags := types.MapValue(types.StringType, elems)
		resp.Diagnostics.Append(diags...)
		state.Settings = val
	}
	if len(secretKeys) > 0 {
		elems := make(map[string]attr.Value, len(secretKeys))
		for _, k := range secretKeys {
			elems[k] = types.StringValue("")
		}
		val, diags := types.MapValue(types.StringType, elems)
		resp.Diagnostics.Append(diags...)
		state.SecretHashes = val
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
