package application

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/adapter"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// DetectProvidersOptions carries parameters for the DetectProviders use case.
type DetectProvidersOptions struct {
	RootPath string
}

// InjectProviderOptions carries parameters for the InjectProvider use case.
type InjectProviderOptions struct {
	RootPath     string
	ProviderName string
}

// DetectProviders returns detection results for all known AI CLI providers
// without making any changes to the filesystem.
func DetectProviders(_ context.Context, opts DetectProvidersOptions) ([]adapter.DetectionResult, error) {
	if _, err := project.Load(opts.RootPath); err != nil {
		return nil, err
	}
	return adapter.Detect(opts.RootPath), nil
}

// InjectProviders injects the Context OS block into all detected providers
// that have not yet been injected. Returns the full detection result list
// with updated Injected flags.
func InjectProviders(_ context.Context, opts DetectProvidersOptions) ([]adapter.DetectionResult, error) {
	if _, err := project.Load(opts.RootPath); err != nil {
		return nil, err
	}

	for _, r := range adapter.Detect(opts.RootPath) {
		if r.Detected && !r.Injected {
			_ = adapter.Inject(opts.RootPath, r.Provider)
		}
	}
	return adapter.Detect(opts.RootPath), nil
}

// InjectProvider injects the Context OS block into a single named provider.
// Works even if the provider binary is not installed (allows manual injection).
func InjectProvider(_ context.Context, opts InjectProviderOptions) error {
	if _, err := project.Load(opts.RootPath); err != nil {
		return err
	}
	for _, p := range adapter.KnownProviders() {
		if p.Name == opts.ProviderName {
			return adapter.Inject(opts.RootPath, p)
		}
	}
	return shared.NewError(shared.CodeNotFound, "unknown provider: "+opts.ProviderName)
}
