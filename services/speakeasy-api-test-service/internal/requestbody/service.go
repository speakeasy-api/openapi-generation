package requestbody

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

func HandleRequestBody(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	var req interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(req); err != nil {
		utils.HandleError(w, err)
	}
}

// FileInfo represents information about an uploaded file
type FileInfo struct {
	FieldName   string `json:"fieldName"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Content     string `json:"content,omitempty"` // Base64 encoded content for small files
}

// MultipartFormResponse represents the response for multipart form uploads
type MultipartFormResponse struct {
	Files      []FileInfo          `json:"files"`
	FormFields map[string][]string `json:"formFields"`
}

// HandleMultipartFormFiles handles multipart form file uploads
func HandleMultipartFormFiles(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with 32MB max memory
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	response := MultipartFormResponse{
		Files:      []FileInfo{},
		FormFields: make(map[string][]string),
	}

	// Handle regular form fields
	for key, values := range r.MultipartForm.Value {
		response.FormFields[key] = values
	}

	// Handle file uploads
	for fieldName, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				utils.HandleError(w, err)
				return
			}
			defer file.Close()

			fileInfo := FileInfo{
				FieldName:   fieldName,
				Filename:    fileHeader.Filename,
				Size:        fileHeader.Size,
				ContentType: fileHeader.Header.Get("Content-Type"),
			}

			// For small files (< 1MB), include base64 encoded content
			if fileHeader.Size < 1024*1024 {
				content, err := io.ReadAll(file)
				if err != nil {
					utils.HandleError(w, err)
					return
				}
				// Store as base64 for JSON compatibility
				fileInfo.Content = base64.StdEncoding.EncodeToString(content)
			}

			response.Files = append(response.Files, fileInfo)
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleError(w, err)
	}
}
