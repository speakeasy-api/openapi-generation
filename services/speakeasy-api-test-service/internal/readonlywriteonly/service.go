package readonlywriteonly

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

type BasicObject struct {
	String string  `json:"string"`
	Bool   bool    `json:"bool"`
	Num    float64 `json:"num"`
}

type WriteOnlyOutputObject struct{}

type InputObject struct {
	Num1 int64 `json:"num1"`
	Num2 int64 `json:"num2"`
	Num3 int64 `json:"num3"`
}

type OutputObject struct {
	Num3 int64 `json:"num3"`
	Sum  int64 `json:"sum"`
}

func HandleReadOrWrite(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	if len(body) > 0 {
		var req BasicObject
		if err := json.Unmarshal(body, &req); err != nil {
			utils.HandleError(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(BasicObject{
		String: "hello",
		Bool:   true,
		Num:    1.0,
	}); err != nil {
		utils.HandleError(w, err)
	}
}

func HandleReadAndWrite(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	var req InputObject
	if err := json.Unmarshal(body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(OutputObject{
		Num3: req.Num3,
		Sum:  req.Num1 + req.Num2 + req.Num3,
	}); err != nil {
		utils.HandleError(w, err)
	}
}

func HandleWriteOnlyOutput(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	var req InputObject
	if err := json.Unmarshal(body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.NewEncoder(w).Encode(WriteOnlyOutputObject{}); err != nil {
		utils.HandleError(w, err)
	}
}
