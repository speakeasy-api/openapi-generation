package oas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// DynamicReservedChecker provides dynamic OpenAPI reserved property checking
type DynamicReservedChecker struct {
	schemaManager *SchemaManager
	cache         map[string]*PropertyContext // version -> extracted properties
	mu            sync.RWMutex
}

// PropertyContext represents where a property can appear and whether it's reserved
type PropertyContext struct {
	PropertyName string
	PathPatterns []string // JSON path patterns where this property is reserved
	Always       bool     // If true, always reserved regardless of context
}

// NewDynamicReservedChecker creates a new dynamic reserved property checker
func NewDynamicReservedChecker() *DynamicReservedChecker {
	return &DynamicReservedChecker{
		schemaManager: NewSchemaManager(),
		cache:         make(map[string]*PropertyContext),
	}
}

// NewDynamicReservedCheckerWithManager creates a checker with a custom schema manager
func NewDynamicReservedCheckerWithManager(schemaManager *SchemaManager) *DynamicReservedChecker {
	return &DynamicReservedChecker{
		schemaManager: schemaManager,
		cache:         make(map[string]*PropertyContext),
	}
}

// IsOASReservedPropertyName checks if a property is reserved in the given context and OpenAPI version
func (drc *DynamicReservedChecker) IsOASReservedPropertyName(propertyName string, path []string, version string) bool {
	if version == "" {
		// No version provided - use hardcoded core properties for basic checking
		return drc.isBasicReservedProperty(propertyName)
	}

	// Check if version is supported before attempting to load schema
	if !drc.IsVersionSupported(version) {
		// For unsupported versions, allow renames (don't restrict)
		// This enables users to rename properties in newer OpenAPI versions
		// until we add support for them
		return false
	}

	// Ensure we have the property contexts for this version
	if err := drc.ensurePropertiesForVersion(version); err != nil {
		// Fallback to basic checking on error (network issues, parsing errors, etc.)
		return drc.isBasicReservedProperty(propertyName)
	}

	return drc.checkReservedProperty(propertyName, path, version)
}

// IsOASReservedPropertyNameSimple provides simple checking for backward compatibility
func (drc *DynamicReservedChecker) IsOASReservedPropertyNameSimple(propertyName string, version string) bool {
	return drc.IsOASReservedPropertyName(propertyName, nil, version)
}

// ensurePropertiesForVersion ensures property contexts are loaded for the given version
func (drc *DynamicReservedChecker) ensurePropertiesForVersion(version string) error {
	cacheKey := version

	drc.mu.RLock()
	_, exists := drc.cache[cacheKey]
	drc.mu.RUnlock()

	if exists {
		return nil
	}

	return drc.loadPropertiesForVersion(version)
}

// loadPropertiesForVersion loads and caches property contexts for a version
func (drc *DynamicReservedChecker) loadPropertiesForVersion(version string) error {
	schema, err := drc.schemaManager.GetSchemaForVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get schema for version %s: %w", version, err)
	}

	// Extract property contexts from schema
	propertyContexts := make(map[string]*PropertyContext)
	drc.extractPropertyContexts(*schema, []string{}, propertyContexts)

	// Cache the property contexts
	drc.mu.Lock()
	for propName, context := range propertyContexts {
		drc.cache[version+"::"+propName] = context
	}
	drc.mu.Unlock()

	return nil
}

// checkReservedProperty checks if a property is reserved using cached contexts
func (drc *DynamicReservedChecker) checkReservedProperty(propertyName string, path []string, version string) bool {
	cacheKey := version + "::" + propertyName

	drc.mu.RLock()
	context, exists := drc.cache[cacheKey]
	drc.mu.RUnlock()

	if !exists {
		return false
	}

	// Check if always reserved
	if context.Always {
		return true
	}

	// Check context-specific restrictions
	if len(context.PathPatterns) > 0 && len(path) > 0 {
		pathStr := strings.Join(path, "/")
		for _, pattern := range context.PathPatterns {
			if strings.Contains(pathStr, pattern) {
				return true
			}
		}
	}

	return false
}

// extractPropertyContexts recursively extracts properties with their path contexts
func (drc *DynamicReservedChecker) extractPropertyContexts(schema Schema, currentPath []string, contexts map[string]*PropertyContext) {
	// Process direct properties
	for propName := range schema.Properties {
		pathStr := strings.Join(currentPath, "/") + "/properties/" + propName

		if contexts[propName] == nil {
			contexts[propName] = &PropertyContext{
				PropertyName: propName,
				PathPatterns: []string{},
			}
		}

		// Determine if this property should be always reserved or context-specific
		if drc.isAlwaysReservedProperty(propName, currentPath) {
			contexts[propName].Always = true
		} else {
			contexts[propName].PathPatterns = append(contexts[propName].PathPatterns, pathStr)
		}
	}

	// Recursively process $defs
	for defName, defSchema := range schema.Defs {
		newPath := append(currentPath, "$defs", defName)
		drc.extractPropertyContexts(defSchema, newPath, contexts)
	}

	// Process schema composition keywords
	for i, subSchema := range schema.AllOf {
		newPath := append(currentPath, "allOf", strconv.Itoa(i))
		drc.extractPropertyContexts(subSchema, newPath, contexts)
	}
	for i, subSchema := range schema.AnyOf {
		newPath := append(currentPath, "anyOf", strconv.Itoa(i))
		drc.extractPropertyContexts(subSchema, newPath, contexts)
	}
	for i, subSchema := range schema.OneOf {
		newPath := append(currentPath, "oneOf", strconv.Itoa(i))
		drc.extractPropertyContexts(subSchema, newPath, contexts)
	}
	if schema.Not != nil {
		newPath := append(currentPath, "not")
		drc.extractPropertyContexts(*schema.Not, newPath, contexts)
	}

	// Process nested schemas in properties
	for propName, prop := range schema.Properties {
		if propMap, ok := prop.(map[string]any); ok {
			propSchemaBytes, err := json.Marshal(propMap)
			if err == nil {
				var propSchema Schema
				if err := json.Unmarshal(propSchemaBytes, &propSchema); err == nil {
					newPath := append(currentPath, "properties", propName)
					drc.extractPropertyContexts(propSchema, newPath, contexts)
				}
			}
		}
	}
}

// isAlwaysReservedProperty determines if a property should always be reserved regardless of context
func (drc *DynamicReservedChecker) isAlwaysReservedProperty(propName string, currentPath []string) bool {
	// Core OpenAPI document properties that should never be renamed
	coreProperties := map[string]bool{
		"openapi": true, "info": true, "servers": true, "paths": true, "webhooks": true,
		"components": true, "security": true, "tags": true, "externalDocs": true,
		"jsonSchemaDialect": true,
	}

	// If this is a top-level property and it's core, always reserve it
	if len(currentPath) <= 1 && coreProperties[propName] {
		return true
	}

	// Schema-specific properties that should be reserved in schema contexts
	schemaProperties := map[string]bool{
		"type": true, "format": true, "enum": true, "const": true, "default": true,
		"multipleOf": true, "maximum": true, "exclusiveMaximum": true, "minimum": true, "exclusiveMinimum": true,
		"maxLength": true, "minLength": true, "pattern": true,
		"maxItems": true, "minItems": true, "uniqueItems": true, "items": true,
		"maxProperties": true, "minProperties": true, "required": true, "properties": true,
		"additionalProperties": true, "description": true, "title": true, "examples": true,
		"readOnly": true, "writeOnly": true, "deprecated": true, "nullable": true,
		"discriminator": true, "xml": true, "externalDocs": true, "example": true,
		"allOf": true, "oneOf": true, "anyOf": true, "not": true, "$ref": true,
	}

	// If we're in a schema definition context, reserve schema properties
	for _, segment := range currentPath {
		if segment == "schemas" || segment == "properties" || segment == "$defs" {
			if schemaProperties[propName] {
				return true
			}
		}
	}

	return false
}

// GetSupportedVersions returns a list of supported OpenAPI versions
func (drc *DynamicReservedChecker) GetSupportedVersions() []string {
	versions := make([]string, 0, len(oasSchemaURLs))
	for version := range oasSchemaURLs {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return versions
}

// IsVersionSupported returns whether the given version is supported
func (drc *DynamicReservedChecker) IsVersionSupported(version string) bool {
	return drc.schemaManager.IsSupported(version)
}

// ClearCache clears the internal property context cache
func (drc *DynamicReservedChecker) ClearCache() {
	drc.mu.Lock()
	drc.cache = make(map[string]*PropertyContext)
	drc.mu.Unlock()
}

// GetCachedVersions returns a list of versions that have been cached
func (drc *DynamicReservedChecker) GetCachedVersions() []string {
	drc.mu.RLock()
	defer drc.mu.RUnlock()

	versionSet := make(map[string]bool)
	for key := range drc.cache {
		if parts := strings.Split(key, "::"); len(parts) > 0 {
			versionSet[parts[0]] = true
		}
	}

	versions := make([]string, 0, len(versionSet))
	for version := range versionSet {
		versions = append(versions, version)
	}
	sort.Strings(versions)
	return versions
}

// isBasicReservedProperty provides basic reserved property checking without version-specific schemas
// Used as fallback when no version is provided or when schema loading fails
func (drc *DynamicReservedChecker) isBasicReservedProperty(propertyName string) bool {
	// Core OpenAPI properties that are always reserved regardless of version
	coreProperties := map[string]bool{
		"openapi": true, "info": true, "servers": true, "paths": true, "webhooks": true,
		"components": true, "security": true, "tags": true, "externalDocs": true,
		"jsonSchemaDialect": true,
		// Common schema properties that are typically reserved
		"$ref": true, "type": true, "format": true, "enum": true, "const": true, "default": true,
		"required": true, "properties": true, "additionalProperties": true, "description": true,
		"title": true, "examples": true, "example": true, "readOnly": true, "writeOnly": true,
		"deprecated": true, "nullable": true, "discriminator": true, "xml": true,
		"allOf": true, "oneOf": true, "anyOf": true, "not": true,
		"multipleOf": true, "maximum": true, "exclusiveMaximum": true, "minimum": true, "exclusiveMinimum": true,
		"maxLength": true, "minLength": true, "pattern": true,
		"maxItems": true, "minItems": true, "uniqueItems": true, "items": true,
		"maxProperties": true, "minProperties": true,
	}

	return coreProperties[propertyName]
}

// Global convenience functions for backward compatibility
// These create temporary dynamic checkers for one-off usage

// IsOASReservedPropertyName provides backward compatibility with the old static function
// For best performance, use DynamicReservedChecker instances in long-running applications like the LSP
func IsOASReservedPropertyName(propertyName string, path []string) bool {
	checker := NewDynamicReservedChecker()
	return checker.IsOASReservedPropertyName(propertyName, path, "")
}

// IsOASReservedPropertyNameSimple provides simple checking for backward compatibility
func IsOASReservedPropertyNameSimple(propertyName string) bool {
	return IsOASReservedPropertyName(propertyName, nil)
}

// IsOASReservedPropertyNameWithVersion provides version-aware checking
func IsOASReservedPropertyNameWithVersion(propertyName string, path []string, version string) bool {
	checker := NewDynamicReservedChecker()
	return checker.IsOASReservedPropertyName(propertyName, path, version)
}
