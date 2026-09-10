package optionalNullableFields

import (
	"encoding/json"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

// MedicalRecord represents a medical record with diagnosis
type MedicalRecord struct {
	Diagnosis string `json:"diagnosis"`
}

// BaseResponse represents the common response structure
type BaseResponse struct {
	ID            int            `json:"id"`
	Type          string         `json:"type,omitempty"`
	MedicalRecord *MedicalRecord `json:"medicalRecord,omitempty"`
}

// RequiredMedicalRecordResponse represents response where medicalRecord is required
type RequiredMedicalRecordResponse struct {
	ID            int            `json:"id"`
	Type          string         `json:"type,omitempty"`
	MedicalRecord *MedicalRecord `json:"medicalRecord"` // No omitempty - field is required
}

// ResponseConfig defines how to handle different scenarios
type ResponseConfig struct {
	UseStructResponse    bool           // If true, use struct response; if false, use map for explicit control
	IncludeMedicalRecord bool           // Whether to include medicalRecord in the response
	MedicalRecordValue   *MedicalRecord // The value to set (nil for null)
}

// EndpointConfig defines the behavior for each endpoint
type EndpointConfig struct {
	IsOptional bool // true if field is optional, false if required
	IsNullable bool // true if field is nullable, false if not nullable
}

// getScenarioConfig returns the configuration for a given scenario
func getScenarioConfig(scenario string) (*ResponseConfig, error) {
	switch scenario {
	case "full":
		return &ResponseConfig{
			UseStructResponse:    true,
			IncludeMedicalRecord: true,
			MedicalRecordValue: &MedicalRecord{
				Diagnosis: "fluffy is sick",
			},
		}, nil
	case "medicalRecordAbsent":
		return &ResponseConfig{
			UseStructResponse:    false, // Use map to control field presence
			IncludeMedicalRecord: false,
			MedicalRecordValue:   nil,
		}, nil
	case "medicalRecordNull":
		return &ResponseConfig{
			UseStructResponse:    false, // Use map to ensure explicit null
			IncludeMedicalRecord: true,
			MedicalRecordValue:   nil,
		}, nil
	default:
		return nil, nil // Invalid scenario
	}
}

// buildResponse creates the appropriate response based on config and endpoint config
func buildResponse(config *ResponseConfig, endpointConfig EndpointConfig) interface{} {
	baseResponse := map[string]interface{}{
		"id":   1,
		"type": "mammal",
	}

	if config.UseStructResponse {
		// Use struct response for clean JSON with omitempty behavior
		if !endpointConfig.IsOptional {
			// Required field - use RequiredMedicalRecordResponse
			return RequiredMedicalRecordResponse{
				ID:            1,
				Type:          "mammal",
				MedicalRecord: config.MedicalRecordValue,
			}
		} else {
			// Optional field - use BaseResponse
			return BaseResponse{
				ID:            1,
				Type:          "mammal",
				MedicalRecord: config.MedicalRecordValue,
			}
		}
	} else {
		// Use map response for explicit control over field presence
		if config.IncludeMedicalRecord {
			baseResponse["medicalRecord"] = config.MedicalRecordValue
		}
		// If IncludeMedicalRecord is false, the field is omitted entirely
		return baseResponse
	}
}

// handleRequest is the unified handler logic for all endpoints
func handleRequest(w http.ResponseWriter, r *http.Request, endpointConfig EndpointConfig) {
	// Get scenario parameter from query string (default to "full")
	scenario := r.URL.Query().Get("scenario")
	if scenario == "" {
		scenario = "full"
	}

	// Get configuration for the scenario
	config, err := getScenarioConfig(scenario)
	if err != nil || config == nil {
		http.Error(w, "Invalid scenario. Use: full, medicalRecordAbsent, or medicalRecordNull", http.StatusBadRequest)
		return
	}

	// Build the response based on configuration
	response := buildResponse(config, endpointConfig)

	// Set headers and encode response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleError(w, err)
	}
}

func HandleObjectWithOptionalTrueNullableTrue(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, EndpointConfig{IsOptional: true, IsNullable: true})
}

func HandleObjectWithOptionalFalseNullableTrue(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, EndpointConfig{IsOptional: false, IsNullable: true})
}

func HandleObjectWithOptionalTrueNullableFalse(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, EndpointConfig{IsOptional: true, IsNullable: false})
}

func HandleObjectWithOptionalFalseNullableFalse(w http.ResponseWriter, r *http.Request) {
	handleRequest(w, r, EndpointConfig{IsOptional: false, IsNullable: false})
}
