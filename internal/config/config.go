package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Port               int      `json:"port"`
	Host               string   `json:"host"`
	RetryAttempts      int      `json:"retry_attempts"`
	RetryDelaySec      int      `json:"retry_delay_sec"`
	RequestTimeoutSec  int      `json:"request_timeout_sec"`
	GeminiBL           string   `json:"gemini_bl"`
	AuthUser           string   `json:"auth_user"`
	XSRFToken          string   `json:"xsrf_token"`
	DefaultModel       string   `json:"default_model"`
	LogRequests        bool     `json:"log_requests"`
	CookieFile         string   `json:"cookie_file"`
	Proxy              string   `json:"proxy"`
	APIKeys            []string `json:"api_keys"`
	Impersonate        string   `json:"impersonate"`
	TemporaryChats     bool     `json:"temporary_chats"`
	UpstreamBase       string   `json:"upstream_base"`
	ImageModel         string   `json:"image_model"`
	ImageBackend       string   `json:"image_backend"`
	LocalImageURL      string   `json:"local_image_url"`
	LocalImageProvider string   `json:"local_image_provider"`
	LocalImageModel    string   `json:"local_image_model"`
	LocalImageKey      string   `json:"local_image_key"`
	ImageModelID       int      `json:"image_model_id"`
	ImageMaxAgeMin     int      `json:"image_max_age_min"`
}

func Default() Config {
	return Config{
		Port:               8081,
		Host:               "0.0.0.0",
		RetryAttempts:      3,
		RetryDelaySec:      2,
		RequestTimeoutSec:  180,
		GeminiBL:           "boq_assistant-bard-web-server_20260716.08_p0",
		AuthUser:           "",
		XSRFToken:          "",
		DefaultModel:       "gemini-3.6-flash",
		LogRequests:        true,
		CookieFile:         "",
		Proxy:              "",
		APIKeys:            []string{},
		Impersonate:        "",
		TemporaryChats:     false,
		UpstreamBase:       "https://gemini.google.com",
		ImageModel:         "gemini-image",
		ImageBackend:       "local",
		LocalImageURL:      "https://image-gen-ebon.vercel.app",
		LocalImageProvider: "alpha",
		LocalImageModel:    "flare",
		LocalImageKey:      "",
		ImageModelID:       7,
		ImageMaxAgeMin:     30,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	var aux struct {
		Port               *int      `json:"port"`
		Host               *string   `json:"host"`
		RetryAttempts      *int      `json:"retry_attempts"`
		RetryDelaySec      *int      `json:"retry_delay_sec"`
		RequestTimeoutSec  *int      `json:"request_timeout_sec"`
		GeminiBL           *string   `json:"gemini_bl"`
		AuthUser           *string   `json:"auth_user"`
		XSRFToken          *string   `json:"xsrf_token"`
		DefaultModel       *string   `json:"default_model"`
		LogRequests        *bool     `json:"log_requests"`
		CookieFile         *string   `json:"cookie_file"`
		Proxy              *string   `json:"proxy"`
		APIKeys            *[]string `json:"api_keys"`
		Impersonate        *string   `json:"impersonate"`
		TemporaryChats     *bool     `json:"temporary_chats"`
		UpstreamBase       *string   `json:"upstream_base"`
		ImageModel         *string   `json:"image_model"`
		ImageBackend       *string   `json:"image_backend"`
		LocalImageURL      *string   `json:"local_image_url"`
		LocalImageProvider *string   `json:"local_image_provider"`
		LocalImageModel    *string   `json:"local_image_model"`
		LocalImageKey      *string   `json:"local_image_key"`
		ImageModelID       *int      `json:"image_model_id"`
		ImageMaxAgeMin     *int      `json:"image_max_age_min"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return cfg, err
	}

	if aux.Port != nil {
		cfg.Port = *aux.Port
	}
	if aux.Host != nil {
		cfg.Host = *aux.Host
	}
	if aux.RetryAttempts != nil {
		cfg.RetryAttempts = *aux.RetryAttempts
	}
	if aux.RetryDelaySec != nil {
		cfg.RetryDelaySec = *aux.RetryDelaySec
	}
	if aux.RequestTimeoutSec != nil {
		cfg.RequestTimeoutSec = *aux.RequestTimeoutSec
	}
	if aux.GeminiBL != nil {
		cfg.GeminiBL = *aux.GeminiBL
	}
	if aux.AuthUser != nil {
		cfg.AuthUser = *aux.AuthUser
	}
	if aux.XSRFToken != nil {
		cfg.XSRFToken = *aux.XSRFToken
	}
	if aux.DefaultModel != nil {
		cfg.DefaultModel = *aux.DefaultModel
	}
	if aux.LogRequests != nil {
		cfg.LogRequests = *aux.LogRequests
	}
	if aux.CookieFile != nil {
		cfg.CookieFile = *aux.CookieFile
	}
	if aux.Proxy != nil {
		cfg.Proxy = *aux.Proxy
	}
	if aux.APIKeys != nil {
		cfg.APIKeys = *aux.APIKeys
	}
	if aux.Impersonate != nil {
		cfg.Impersonate = *aux.Impersonate
	}
	if aux.TemporaryChats != nil {
		cfg.TemporaryChats = *aux.TemporaryChats
	}
	if aux.UpstreamBase != nil {
		cfg.UpstreamBase = *aux.UpstreamBase
	}
	if aux.ImageModel != nil {
		cfg.ImageModel = *aux.ImageModel
	}
	if aux.ImageBackend != nil {
		cfg.ImageBackend = *aux.ImageBackend
	}
	if aux.LocalImageURL != nil {
		cfg.LocalImageURL = *aux.LocalImageURL
	}
	if aux.LocalImageProvider != nil {
		cfg.LocalImageProvider = *aux.LocalImageProvider
	}
	if aux.LocalImageModel != nil {
		cfg.LocalImageModel = *aux.LocalImageModel
	}
	if aux.LocalImageKey != nil {
		cfg.LocalImageKey = *aux.LocalImageKey
	}
	if aux.ImageModelID != nil {
		cfg.ImageModelID = *aux.ImageModelID
	}
	if aux.ImageMaxAgeMin != nil {
		cfg.ImageMaxAgeMin = *aux.ImageMaxAgeMin
	}

	return cfg, nil
}

func Find() string {
	paths := []string{"./config.json"}
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "gemini-web2api", "config.json"))
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
