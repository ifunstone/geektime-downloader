package uiapp

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type persistedSettings struct {
	Gcid                   string `json:"gcid"`
	Gcess                  string `json:"gcess"`
	DownloadFolder         string `json:"download_folder"`
	Quality                string `json:"quality"`
	DownloadComments       string `json:"download_comments"`
	PrintPDFWaitSeconds    string `json:"print_pdf_wait_seconds"`
	PrintPDFTimeoutSeconds string `json:"print_pdf_timeout_seconds"`
	Interval               string `json:"interval"`
	IsEnterprise           bool   `json:"is_enterprise"`
	LogLevel               string `json:"log_level"`
	OutputPDF              *bool  `json:"output_pdf"`
	OutputMarkdown         *bool  `json:"output_markdown"`
	OutputAudio            *bool  `json:"output_audio"`
	LastProductID          string `json:"last_product_id"`
	LastProductType        string `json:"last_product_type"`
}

func settingsFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "geektime-downloader")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "ui-settings.json"), nil
}

func loadSettings() (*persistedSettings, error) {
	path, err := settingsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &persistedSettings{}, nil
		}
		return nil, err
	}
	var settings persistedSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func saveSettings(settings persistedSettings) error {
	path, err := settingsFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
