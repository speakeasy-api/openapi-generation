package oas

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SchemaManager handles dynamic downloading and caching of OpenAPI schemas
type SchemaManager struct {
	cache     map[string]*Schema        // version -> parsed schema
	httpCache map[string]httpCacheEntry // URL -> cached HTTP response
	client    *http.Client
	config    SchemaManagerConfig
	mu        sync.RWMutex
}

// SchemaManagerConfig provides configuration options for the schema manager
type SchemaManagerConfig struct {
	DisableRemoteSchemas bool              // Force offline mode
	CustomSchemaURLs     map[string]string // Override default URLs
	CacheTTL             time.Duration     // HTTP cache TTL
	HTTPTimeout          time.Duration     // Network request timeout
	MaxCacheSize         int               // Maximum number of cached schemas
}

// httpCacheEntry represents a cached HTTP response
type httpCacheEntry struct {
	data         []byte
	timestamp    time.Time
	ttl          time.Duration
	etag         string
	lastModified string
}

// Schema represents a parsed JSON Schema object
type Schema struct {
	Properties           map[string]any    `json:"properties"`
	PatternProperties    map[string]any    `json:"patternProperties"`
	Defs                 map[string]Schema `json:"$defs"`
	Items                any               `json:"items"`
	AdditionalProperties any               `json:"additionalProperties"`
	AllOf                []Schema          `json:"allOf"`
	AnyOf                []Schema          `json:"anyOf"`
	OneOf                []Schema          `json:"oneOf"`
	Not                  *Schema           `json:"not"`
	Ref                  string            `json:"$ref"`
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

// DefaultSchemaManagerConfig returns the default configuration
func DefaultSchemaManagerConfig() SchemaManagerConfig {
	return SchemaManagerConfig{
		DisableRemoteSchemas: false,
		CustomSchemaURLs:     make(map[string]string),
		CacheTTL:             24 * time.Hour,
		HTTPTimeout:          10 * time.Second,
		MaxCacheSize:         10,
	}
}

// NewSchemaManager creates a new schema manager with default configuration
func NewSchemaManager() *SchemaManager {
	return NewSchemaManagerWithConfig(DefaultSchemaManagerConfig())
}

// NewSchemaManagerWithConfig creates a new schema manager with custom configuration
func NewSchemaManagerWithConfig(config SchemaManagerConfig) *SchemaManager {
	return &SchemaManager{
		cache:     make(map[string]*Schema),
		httpCache: make(map[string]httpCacheEntry),
		client: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		config: config,
	}
}

// OpenAPI schema URLs mapping
var oasSchemaURLs = map[string]string{
	"3.1.0": "https://spec.openapis.org/oas/3.1/schema/2025-02-13",
	"3.1.1": "https://spec.openapis.org/oas/3.1/schema/2025-02-13",
	"3.0.3": "https://spec.openapis.org/oas/3.0/schema/2021-09-28",
	"3.0.2": "https://spec.openapis.org/oas/3.0/schema/2021-09-28",
	"3.0.1": "https://spec.openapis.org/oas/3.0/schema/2021-09-28",
	"3.0.0": "https://spec.openapis.org/oas/3.0/schema/2021-09-28",
}

// GetSchemaForVersion returns the parsed schema for the given OpenAPI version
func (sm *SchemaManager) GetSchemaForVersion(version string) (*Schema, error) {
	if version == "" {
		return nil, errors.New("empty version provided")
	}

	// Normalize version (remove pre-release identifiers)
	normalizedVersion := normalizeVersion(version)

	sm.mu.RLock()
	if schema, exists := sm.cache[normalizedVersion]; exists {
		sm.mu.RUnlock()
		return schema, nil
	}
	sm.mu.RUnlock()

	// Get schema URL
	url, err := sm.getSchemaURLForVersion(normalizedVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema URL for version %s: %w", version, err)
	}

	// Download and parse schema
	schema, err := sm.fetchAndParseSchema(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema for version %s: %w", version, err)
	}

	// Cache the schema
	sm.cacheSchema(normalizedVersion, schema)
	return schema, nil
}

// getSchemaURLForVersion returns the schema URL for a given version
func (sm *SchemaManager) getSchemaURLForVersion(version string) (string, error) {
	// Check custom URLs first
	if url, exists := sm.config.CustomSchemaURLs[version]; exists {
		return url, nil
	}

	// Check default URLs
	if url, exists := oasSchemaURLs[version]; exists {
		return url, nil
	}

	return "", fmt.Errorf("unsupported OpenAPI version: %s", version)
}

// fetchAndParseSchema downloads and parses a schema from the given URL
func (sm *SchemaManager) fetchAndParseSchema(url string) (*Schema, error) {
	if sm.config.DisableRemoteSchemas {
		return nil, errors.New("remote schemas disabled")
	}

	// Try to get from HTTP cache first
	if cachedData, err := sm.getCachedHTTPResponse(url); err == nil {
		return sm.parseSchema(cachedData)
	}

	// Download from remote
	data, err := sm.downloadSchema(url)
	if err != nil {
		return nil, err
	}

	// Parse the schema
	schema, err := sm.parseSchema(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Cache the HTTP response
	sm.cacheHTTPResponse(url, data)

	return schema, nil
}

// downloadSchema downloads a schema from the given URL
func (sm *SchemaManager) downloadSchema(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add cache headers if we have cached data
	sm.mu.RLock()
	if cached, exists := sm.httpCache[url]; exists {
		if cached.etag != "" {
			req.Header.Set("If-None-Match", cached.etag)
		}
		if cached.lastModified != "" {
			req.Header.Set("If-Modified-Since", cached.lastModified)
		}
	}
	sm.mu.RUnlock()

	resp, err := sm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download schema: %w", err)
	}
	defer resp.Body.Close()

	// Handle 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		sm.mu.RLock()
		cached := sm.httpCache[url]
		sm.mu.RUnlock()
		return cached.data, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	var data []byte
	if resp.ContentLength > 0 {
		data = make([]byte, 0, resp.ContentLength)
	} else {
		data = make([]byte, 0)
	}
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return data, nil
}

// parseSchema parses JSON schema data into a Schema struct
func (sm *SchemaManager) parseSchema(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema JSON: %w", err)
	}
	return &schema, nil
}

// cacheSchema stores a parsed schema in memory
func (sm *SchemaManager) cacheSchema(version string, schema *Schema) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Simple LRU eviction if cache is full
	if len(sm.cache) >= sm.config.MaxCacheSize {
		// Remove a random entry (simple eviction strategy)
		for k := range sm.cache {
			delete(sm.cache, k)
			break
		}
	}

	sm.cache[version] = schema
}

// cacheHTTPResponse stores an HTTP response in the cache
func (sm *SchemaManager) cacheHTTPResponse(url string, data []byte) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.httpCache[url] = httpCacheEntry{
		data:      data,
		timestamp: time.Now(),
		ttl:       sm.config.CacheTTL,
	}
}

// getCachedHTTPResponse retrieves a cached HTTP response if valid
func (sm *SchemaManager) getCachedHTTPResponse(url string) ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	cached, exists := sm.httpCache[url]
	if !exists {
		return nil, errors.New("no cached response")
	}

	if time.Since(cached.timestamp) > cached.ttl {
		return nil, errors.New("cached response expired")
	}

	return cached.data, nil
}

// normalizeVersion normalizes a version string by removing pre-release identifiers
func normalizeVersion(version string) string {
	// Remove pre-release identifiers (e.g., "3.1.0-beta" -> "3.1.0")
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}
	if idx := strings.Index(version, "+"); idx != -1 {
		version = version[:idx]
	}
	return version
}

// IsSupported returns whether the given version is supported
func (sm *SchemaManager) IsSupported(version string) bool {
	normalizedVersion := normalizeVersion(version)

	// Check custom URLs
	if _, exists := sm.config.CustomSchemaURLs[normalizedVersion]; exists {
		return true
	}

	// Check default URLs
	_, exists := oasSchemaURLs[normalizedVersion]
	return exists
}
