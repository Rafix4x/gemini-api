package models

import (
	"fmt"
	"strconv"
	"strings"
)

type Model struct {
	Mode  int         `json:"mode"`
	Think int         `json:"think"`
	Desc  string      `json:"desc"`
	Extra map[int]any `json:"extra,omitempty"`
	Image bool        `json:"image,omitempty"`
}

// Anonymous upstream responses reported "3.5 Flash-Lite" (mode 6) for every
// tested chat selector. Expose that single route instead of invented versions.
const DefaultModelName = "gemini-3.5-flash-lite"
const UnknownUpstreamModel = "gemini-web"

var MODELS = map[string]Model{
	DefaultModelName: {Mode: 6, Think: 4, Desc: "Gemini web chat; anonymous upstream metadata reported 3.5 Flash-Lite. Account routing can change."},
}

type Resolved struct {
	Name  string
	Mode  int
	Think int
	Extra map[int]any
}

func Resolve(modelName, defaultName string) (Resolved, error) {
	if modelName == "" {
		modelName = defaultName
		if modelName == "" {
			modelName = DefaultModelName
		}
	}
	var thinkOverride *int
	if idx := strings.LastIndex(modelName, "@think="); idx != -1 {
		value, err := strconv.Atoi(modelName[idx+len("@think="):])
		if err != nil || value < 0 || value > 4 {
			return Resolved{}, fmt.Errorf("think level must be an integer from 0 to 4")
		}
		modelName = modelName[:idx]
		thinkOverride = &value
	}
	cfg, exists := MODELS[modelName]
	if !exists {
		return Resolved{}, fmt.Errorf("unknown chat model %q; use %q", modelName, DefaultModelName)
	}
	think := cfg.Think
	if thinkOverride != nil {
		think = *thinkOverride
	}
	return Resolved{Name: modelName, Mode: cfg.Mode, Think: think, Extra: cfg.Extra}, nil
}

// ResponseName uses the model label from upstream metadata, never model prose
// or the name requested by the caller. Missing metadata stays explicitly unknown.
func ResponseName(label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	if label == "" {
		return UnknownUpstreamModel
	}
	label = strings.TrimPrefix(label, "gemini ")
	label = strings.ReplaceAll(label, " ", "-")
	if !strings.HasPrefix(label, "gemini-") {
		label = "gemini-" + label
	}
	if len(label) > 100 {
		return UnknownUpstreamModel
	}
	for _, char := range label {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-' || char == '.') {
			return UnknownUpstreamModel
		}
	}
	return label
}
