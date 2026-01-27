package config

//go:generate go run ./gen

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var Config = field.NewConfiguration([]field.SchemaField{
	field.StringField(
		"api-key",
		field.WithRequired(true),
		field.WithDescription("PandaDoc account API-Key"),
	),
	field.BoolField(
		"europe-domain",
		field.WithDescription("PandaDoc API domain. Set to true for europe domain (false by default)."),
		field.WithDefaultValue(false),
	),
})

func ValidateConfig(c *PandaDoc) error {
	return nil
}
