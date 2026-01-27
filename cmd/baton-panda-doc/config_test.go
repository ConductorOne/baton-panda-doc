package main

import (
	"testing"

	"github.com/conductorone/baton-panda-doc/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/test"
)

func TestConfigs(t *testing.T) {
	configurationSchema := field.NewConfiguration(
		config.ConfigurationFields,
		field.WithConstraints(config.FieldRelationships...),
	)

	test.ExerciseTestCases(t, configurationSchema, ValidateConfig, []test.TestCase{})
}
