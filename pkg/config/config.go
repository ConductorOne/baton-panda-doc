package config

//go:generate go run ./gen

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var Config = field.NewConfiguration([]field.SchemaField{
	field.StringField(
		"api-key",
		field.WithRequired(true),
		field.WithIsSecret(true),
		field.WithDescription("PandaDoc account API-Key"),
		field.WithDisplayName("API Key"),
		field.WithPlaceholder("Enter your PandaDoc API key"),
	),
	field.BoolField(
		"europe-domain",
		field.WithDisplayName("Use PandaDoc Europe domain"),
		field.WithDescription("PandaDoc API domain. Set to true for europe domain (false by default)."),
		field.WithDefaultValue(false),
	),
	field.StringField(
		"base-url",
		field.WithDescription("Override the PandaDoc API URL (for testing)"),
		field.WithHidden(true),
		field.WithExportTarget(field.ExportTargetCLIOnly),
	),
},
	field.WithConnectorDisplayName("PandaDoc"),
	field.WithIconUrl("/static/app-icons/panda-doc.svg"),
	field.WithHelpUrl("/docs/baton/panda-doc"),
)

func ValidateConfig(c *PandaDoc) error {
	return nil
}
