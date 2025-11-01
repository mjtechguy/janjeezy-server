package main

import (
	"context"
	"fmt"

	"menlo.ai/indigo-api-gateway/app/domain/auth"
	"menlo.ai/indigo-api-gateway/app/domain/model"
	"menlo.ai/indigo-api-gateway/app/domain/organization"
	"menlo.ai/indigo-api-gateway/app/domain/settings"
	"menlo.ai/indigo-api-gateway/app/infrastructure/inference"
	"menlo.ai/indigo-api-gateway/config/environment_variables"
)

type DataInitializer struct {
	authService         *auth.AuthService
	providerRegistry    *model.ProviderRegistryService
	modelCatalogService *model.ModelCatalogService
	inferenceProvider   *inference.InferenceProvider
	settingsService     *settings.Service
}

func (d *DataInitializer) Install(ctx context.Context) error {
	err := d.installDefaultOrganization(ctx)
	if err != nil {
		return err
	}

	if err := d.applyStoredSettings(ctx); err != nil {
		return err
	}

	if environment_variables.EnvironmentVariables.JAN_INFERENCE_SETUP {
		err = d.setupJanProvider(ctx)
		if err != nil {
			return fmt.Errorf("failed to setup Jan provider: %v", err)
		}
	}

	return nil
}

func (d *DataInitializer) installDefaultOrganization(ctx context.Context) error {
	return d.authService.InitOrganization(ctx)
}

func (d *DataInitializer) applyStoredSettings(ctx context.Context) error {
	if organization.DEFAULT_ORGANIZATION == nil {
		return nil
	}
	smtp, err := d.settingsService.GetSMTPSettings(ctx, organization.DEFAULT_ORGANIZATION.ID)
	if err != nil {
		return err
	}
	if smtp != nil && smtp.Host != "" {
		env := &environment_variables.EnvironmentVariables
		env.SMTP_HOST = smtp.Host
		env.SMTP_PORT = smtp.Port
		env.SMTP_USERNAME = smtp.Username
		if smtp.Password != "" {
			env.SMTP_PASSWORD = smtp.Password
		}
		env.SMTP_SENDER_EMAIL = smtp.FromEmail
	}
	return nil
}

func (d *DataInitializer) setupJanProvider(ctx context.Context) error {
	// Skip if default organization is not set
	if organization.DEFAULT_ORGANIZATION == nil {
		return fmt.Errorf("default organization not initialized")
	}

	// Check if Jan provider already exists for default org (organization-scoped providers)
	providers, err := d.providerRegistry.ListAccessibleProviders(ctx, organization.DEFAULT_ORGANIZATION.ID, nil)
	if err != nil {
		return err
	}
	for _, p := range providers {
		if p == nil {
			continue
		}
		// only consider organization-scoped providers (no project)
		if p.Kind == model.ProviderJan && p.ProjectID == nil {
			return nil
		}
	}

	// Create new Jan provider for default organization
	result, regErr := d.providerRegistry.RegisterProvider(ctx, model.RegisterProviderInput{
		OrganizationID: organization.DEFAULT_ORGANIZATION.ID,
		Name:           "Jan Shared Key",
		Vendor:         string(model.ProviderJan),
		BaseURL:        environment_variables.EnvironmentVariables.JAN_INFERENCE_MODEL_URL,
		APIKey:         "none",
		Metadata: map[string]string{
			"description": "Default organization access to Jan Provider",
		},
		Active: true,
	})
	if regErr != nil {
		// RegisterProvider returns *common.Error. Convert to error with message.
		return fmt.Errorf("register provider failed: %v", regErr)
	}

	models, err := d.inferenceProvider.ListModels(ctx, result.Provider)
	if err != nil {
		return fmt.Errorf("failed to discover models for jan provider: %w", err)
	}
	if _, syncErr := d.providerRegistry.SyncProviderModels(ctx, result.Provider, models); syncErr != nil {
		return fmt.Errorf("failed to sync jan provider models: %v", syncErr)
	}

	return nil
}
