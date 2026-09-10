package ast

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	generationtelemetry "github.com/speakeasy-api/generation-context/telemetry"
	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/naming"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"golang.org/x/sync/errgroup"
)

// TerraformGoTypeName sanitizes a name for use as a Go type name by applying
// SanitizeName and converting to GoPascal case with acronym preservation. Used
// for both entity type names and symbol names in generated Terraform provider
// code.
func TerraformGoTypeName(name string) string {
	return casing.New().ToGoPascal(sanitization.SanitizeName(name))
}

// Describes all Terraform Provider related data in the AST as collected from
// various x-speakeasy-entity* extensions. Data is split by resource type.
type TerraformProvider struct {
	// Valid actions sorted by name. Populated by AssembleSchemas.
	//
	// Terraform actions are entities with an invoke lifecycle operation.
	// These are represented in Terraform using "action" configuration blocks.
	Actions []*TerraformAction `json:"actions" yaml:"actions"`

	// Valid data resources sorted by name. Populated by AssembleSchemas.
	//
	// Terraform data resources are entities with only a Read lifecycle
	// operation. These are represented in Terraform using "data" configuration
	// blocks and colloquially referred to as "data sources".
	DataResources []*TerraformDataResource `json:"data_resources" yaml:"data_resources"`

	// Valid ephemeral resources sorted by name. Populated by AssembleSchemas.
	//
	// Terraform ephemeral resources are entities with an open lifecycle
	// operation. These are represented in Terraform using "ephemeral"
	// configuration blocks.
	EphemeralResources []*TerraformEphemeralResource `json:"ephemeral_resources" yaml:"ephemeral_resources"`

	// Valid managed resources sorted by name. Populated by AssembleSchemas.
	//
	// Terraform managed resources are entities with Create, Read, Update,
	// and Delete lifecycle operations. These are represented in Terraform
	// using "resource" configuration blocks and colloquially referred to as
	// "resources" due to that implementation detail and existing before other
	// resource types in Terraform.
	ManagedResources []*TerraformManagedResource `json:"managed_resources" yaml:"managed_resources"`

	// Resolved Terraform provider type name (e.g. "myprovider"). Used as the
	// value of resp.TypeName in the provider Metadata implementation, and as
	// the prefix for all resource, data source, ephemeral resource, and
	// action type names. Sourced from the providerTypeNameOverride generation
	// configuration if set, otherwise from packageName. Must be set before
	// any AddOrGet* method is called so that entity TerraformTypeName fields
	// are composed with the correct prefix.
	TerraformTypeName string `json:"terraformTypeName" yaml:"terraformTypeName"`

	// Raw actions keyed by entity name. Populated by AddOrGetAction.
	actions map[string]*TerraformAction

	// Raw data resources keyed by entity name. Populated by AddOrGetDataResource.
	dataResources map[string]*TerraformDataResource

	// Raw ephemeral resources keyed by entity name. Populated by
	// AddOrGetEphemeralResource.
	ephemeralResources map[string]*TerraformEphemeralResource

	// Ensures missing telemetry event logging only happens once.
	logMissingTelemetryEventOnce sync.Once

	// Raw managed resources keyed by entity name. Populated by
	// AddOrGetManagedResource.
	managedResources map[string]*TerraformManagedResource
}

// Creates a new Terraform AST, safely initializing underlying fields.
func NewTerraformProvider() *TerraformProvider {
	return &TerraformProvider{
		actions:            make(map[string]*TerraformAction),
		dataResources:      make(map[string]*TerraformDataResource),
		ephemeralResources: make(map[string]*TerraformEphemeralResource),
		managedResources:   make(map[string]*TerraformManagedResource),
	}
}

// Adds a new action or retrieves an existing action.
func (a *TerraformProvider) AddOrGetAction(name string) (*TerraformAction, error) {
	if !terraform.IsResourceNameValid(name) {
		return nil, fmt.Errorf("invalid action name: %s", name)
	}

	if a.actions == nil {
		a.actions = make(map[string]*TerraformAction)
	}

	if action, ok := a.actions[name]; ok {
		return action, nil
	}

	result := NewTerraformAction(name)
	result.TerraformTypeName = terraform.TypeName(a.TerraformTypeName, name)

	a.actions[name] = result

	return result, nil
}

// Adds a new data resource or retrieves an existing data resource.
func (a *TerraformProvider) AddOrGetDataResource(name string) (*TerraformDataResource, error) {
	if !terraform.IsResourceNameValid(name) {
		return nil, fmt.Errorf("invalid data resource name: %s", name)
	}

	if a.dataResources == nil {
		a.dataResources = make(map[string]*TerraformDataResource)
	}

	if dataResource, ok := a.dataResources[name]; ok {
		return dataResource, nil
	}

	result := NewTerraformDataResource(name)
	result.TerraformTypeName = terraform.TypeName(a.TerraformTypeName, name)

	a.dataResources[name] = result

	return result, nil
}

// Adds a new ephemeral resource or retrieves an existing ephemeral resource.
func (a *TerraformProvider) AddOrGetEphemeralResource(name string) (*TerraformEphemeralResource, error) {
	if !terraform.IsResourceNameValid(name) {
		return nil, fmt.Errorf("invalid ephemeral resource name: %s", name)
	}

	if a.ephemeralResources == nil {
		a.ephemeralResources = make(map[string]*TerraformEphemeralResource)
	}

	if ephemeralResource, ok := a.ephemeralResources[name]; ok {
		return ephemeralResource, nil
	}

	result := NewTerraformEphemeralResource(name)
	result.TerraformTypeName = terraform.TypeName(a.TerraformTypeName, name)

	a.ephemeralResources[name] = result

	return result, nil
}

// Adds a new managed resource or retrieves an existing managed resource.
func (a *TerraformProvider) AddOrGetManagedResource(name string) (*TerraformManagedResource, error) {
	if !terraform.IsResourceNameValid(name) {
		return nil, fmt.Errorf("invalid managed resource name: %s", name)
	}

	if a.managedResources == nil {
		a.managedResources = make(map[string]*TerraformManagedResource)
	}

	if managedResource, ok := a.managedResources[name]; ok {
		return managedResource, nil
	}

	result := NewTerraformManagedResource(name)
	result.TerraformTypeName = terraform.TypeName(a.TerraformTypeName, name)

	a.managedResources[name] = result

	return result, nil
}

// Adds an operation to the appropriate resources based on the
// x-speakeasy-entity-operation configuration.
func (p *TerraformProvider) AddOperation(generationConfig map[string]any, operation *Operation) error {
	if operation == nil {
		return nil
	}

	entityOperation := operation.Extensions.EntityOperation

	for _, entityOperationConfig := range entityOperation.TerraformActions {
		action, err := p.AddOrGetAction(entityOperationConfig.Entity)

		if err != nil {
			return fmt.Errorf("failed to add Terraform action %q: %w", entityOperationConfig.Entity, err)
		}

		if err := action.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
			return fmt.Errorf("failed to add Terraform action %q operation %q: %w", entityOperationConfig.Entity, operation.ID, err)
		}
	}

	for _, entityOperationConfig := range entityOperation.TerraformDataResources {
		dataResource, err := p.AddOrGetDataResource(entityOperationConfig.Entity)

		if err != nil {
			return fmt.Errorf("failed to add Terraform data resource %q: %w", entityOperationConfig.Entity, err)
		}

		if err := dataResource.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
			return fmt.Errorf("failed to add Terraform data resource %q operation %q: %w", entityOperationConfig.Entity, operation.ID, err)
		}
	}

	for _, entityOperationConfig := range entityOperation.TerraformEphemeralResources {
		ephemeralResource, err := p.AddOrGetEphemeralResource(entityOperationConfig.Entity)

		if err != nil {
			return fmt.Errorf("failed to add Terraform ephemeral resource %q: %w", entityOperationConfig.Entity, err)
		}

		if err := ephemeralResource.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
			return fmt.Errorf("failed to add Terraform ephemeral resource %q operation %q: %w", entityOperationConfig.Entity, operation.ID, err)
		}
	}

	for _, entityOperationConfig := range entityOperation.TerraformManagedResources {
		managedResource, err := p.AddOrGetManagedResource(entityOperationConfig.Entity)

		if err != nil {
			return fmt.Errorf("failed to add Terraform managed resource %q: %w", entityOperationConfig.Entity, err)
		}

		if err := managedResource.AddOperation(generationConfig, entityOperationConfig, operation); err != nil {
			return fmt.Errorf("failed to add Terraform managed resource %q operation %q: %w", entityOperationConfig.Entity, operation.ID, err)
		}
	}

	return nil
}

// AssembleSchemas calls AssembleSchemaTypeDef on all valid entities in the
// provider. This must be called after all operations have been added. Schema
// warnings from managed resources are logged as warnings.
func (p *TerraformProvider) AssembleSchemas(ctx context.Context, excludeEmptyObjectSchemas bool) error {
	log := logging.From(ctx)
	start := time.Now()

	// Pre-filter valid entities in sorted order for each type.
	var validActions []*TerraformAction
	for _, name := range slices.Sorted(maps.Keys(p.actions)) {
		if a := p.actions[name]; a.isValid() {
			validActions = append(validActions, a)
		}
	}

	var validDataResources []*TerraformDataResource
	for _, name := range slices.Sorted(maps.Keys(p.dataResources)) {
		if dr := p.dataResources[name]; dr.isValid() {
			validDataResources = append(validDataResources, dr)
		}
	}

	var validEphemeralResources []*TerraformEphemeralResource
	for _, name := range slices.Sorted(maps.Keys(p.ephemeralResources)) {
		if er := p.ephemeralResources[name]; er.isValid() {
			validEphemeralResources = append(validEphemeralResources, er)
		}
	}

	var validManagedResources []*TerraformManagedResource
	for _, name := range slices.Sorted(maps.Keys(p.managedResources)) {
		if mr := p.managedResources[name]; mr.isValid() {
			validManagedResources = append(validManagedResources, mr)
		}
	}

	// Pre-allocate result slices to enable index-based writes without a mutex.
	p.Actions = make([]*TerraformAction, len(validActions))
	p.DataResources = make([]*TerraformDataResource, len(validDataResources))
	p.EphemeralResources = make([]*TerraformEphemeralResource, len(validEphemeralResources))
	p.ManagedResources = make([]*TerraformManagedResource, len(validManagedResources))

	// Assemble all entity schemas concurrently. Each AssembleSchemaTypeDef
	// call operates only on its own receiver data (shards were cloned during
	// operation setup). Shared APIOperation pointers are read-only during
	// assembly.
	var g errgroup.Group

	for i, action := range validActions {
		g.Go(func() error {
			resourceStart := time.Now()

			if err := action.AssembleSchemaTypeDef(excludeEmptyObjectSchemas); err != nil {
				return err
			}

			p.Actions[i] = action

			log.Debug(fmt.Sprintf("assembled action %q schema in %s", action.Name, time.Since(resourceStart)))

			return nil
		})
	}

	for i, dataResource := range validDataResources {
		g.Go(func() error {
			resourceStart := time.Now()

			if err := dataResource.AssembleSchemaTypeDef(excludeEmptyObjectSchemas); err != nil {
				return err
			}

			p.DataResources[i] = dataResource

			log.Debug(fmt.Sprintf("assembled data resource %q schema in %s", dataResource.Name, time.Since(resourceStart)))

			return nil
		})
	}

	for i, ephemeralResource := range validEphemeralResources {
		g.Go(func() error {
			resourceStart := time.Now()

			if err := ephemeralResource.AssembleSchemaTypeDef(excludeEmptyObjectSchemas); err != nil {
				return err
			}

			p.EphemeralResources[i] = ephemeralResource

			log.Debug(fmt.Sprintf("assembled ephemeral resource %q schema in %s", ephemeralResource.Name, time.Since(resourceStart)))

			return nil
		})
	}

	for i, managedResource := range validManagedResources {
		g.Go(func() error {
			resourceStart := time.Now()

			if err := managedResource.AssembleSchemaTypeDef(excludeEmptyObjectSchemas); err != nil {
				return err
			}

			p.ManagedResources[i] = managedResource

			log.Debug(fmt.Sprintf("assembled managed resource %q schema in %s", managedResource.Name, time.Since(resourceStart)))

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Log managed resource schema warnings after all assemblies complete.
	for _, managedResource := range p.ManagedResources {
		for _, warning := range managedResource.SchemaWarnings {
			log.Warn(warning)
		}
	}

	log.Debug(fmt.Sprintf("assembled all schemas in %s", time.Since(start)))

	p.updateTelemetryEvent(ctx)

	return nil
}

// AnnotateTerraformSymbols assigns unique deduplicated "Symbol" extension
// values to all complex types (class and union) across all entity
// SchemaTypeDefs. Structurally identical types may share a symbol to reduce
// generated code duplication. When enableTypeDeduplication is true, any
// structurally matching type is deduplicated regardless of name; when false,
// types must also share the same candidate name. Must be called after
// AssembleSchemas.
func (p *TerraformProvider) AnnotateTerraformSymbols(ctx context.Context, enableTypeDeduplication bool) error {
	log := logging.From(ctx)
	start := time.Now()

	initialSymbols := make(map[string]*TypeDef, len(naming.GoReservedNames))
	for name := range naming.GoReservedNames {
		initialSymbols[TerraformGoTypeName(name)] = nil
	}

	sm := newSymbolManager(initialSymbols, enableTypeDeduplication)

	// Walk entities in order: Actions, ManagedResources, DataResources,
	// EphemeralResources. Each group is pre-sorted by AssembleSchemas.
	for _, action := range p.Actions {
		if action.SchemaTypeDef == nil {
			continue
		}

		if err := action.SchemaTypeDef.annotateTerraformSymbols(sm); err != nil {
			return err
		}
	}

	for _, managedResource := range p.ManagedResources {
		if managedResource.SchemaTypeDef == nil {
			continue
		}

		if err := managedResource.SchemaTypeDef.annotateTerraformSymbols(sm); err != nil {
			return err
		}
	}

	for _, dataResource := range p.DataResources {
		if dataResource.SchemaTypeDef == nil {
			continue
		}

		if err := dataResource.SchemaTypeDef.annotateTerraformSymbols(sm); err != nil {
			return err
		}
	}

	for _, ephemeralResource := range p.EphemeralResources {
		if ephemeralResource.SchemaTypeDef == nil {
			continue
		}

		if err := ephemeralResource.SchemaTypeDef.annotateTerraformSymbols(sm); err != nil {
			return err
		}
	}

	log.Debug(fmt.Sprintf("annotated all terraform symbols in %s", time.Since(start)))

	return nil
}

// Updates the telemetry event with the TerraformProvider resource counts. These
// counts are used for billing/licensing purposes. Must be called after
// AssembleSchemas populates the exported slice fields.
func (p *TerraformProvider) updateTelemetryEvent(ctx context.Context) {
	event := generationtelemetry.EventFromContext(ctx)

	// The event may be nil in certain contexts, such as generator testing
	// targets. Only log once.
	if event == nil {
		p.logMissingTelemetryEventOnce.Do(func() {
			logging.From(ctx).Debug("failed to get telemetry event for Terraform Provider update")
		})
		return
	}

	// Track number of actions, ephemeral resources, and managed resources for
	// billing/licensing purposes. In the future, other resource types or
	// provider concepts may be tracked for billing as well, such as list
	// resource types.
	count := int64(len(p.Actions) + len(p.EphemeralResources) + len(p.ManagedResources))
	event.GenerateNumberOfTerraformResources = &count
}
