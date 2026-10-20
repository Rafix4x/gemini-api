package server

import (
	"net/http"

	"github.com/Rafix4x/gemini-api/internal/models"
)

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	modelNames := make([]string, 0, len(models.MODELS))
	for k := range models.MODELS {
		modelNames = append(modelNames, k)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": a.Version,
		"models":  modelNames,
	})
}

func (a *App) handleModels(w http.ResponseWriter, r *http.Request) {
	var data []map[string]any
	for name, m := range models.MODELS {
		entry := map[string]any{
			"id":          name,
			"object":      "model",
			"created":     1700000000,
			"owned_by":    "google",
			"description": m.Desc,
		}
		if m.Image {
			entry["capabilities"] = []string{"image_generation"}
		}
		data = append(data, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"object": "list",
		"data":   data,
	})
}

func (a *App) handleGoogleModels(w http.ResponseWriter, r *http.Request) {
	var modelList []map[string]any
	for name, m := range models.MODELS {
		methods := []string{"generateContent", "streamGenerateContent"}
		if m.Image {
			methods = append(methods, "predict")
		}
		modelList = append(modelList, map[string]any{
			"name":                       "models/" + name,
			"displayName":                name,
			"description":                m.Desc,
			"supportedGenerationMethods": methods,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"models": modelList,
	})
}
