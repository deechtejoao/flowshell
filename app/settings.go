package app

import (
	"encoding/json"
	"os"
)

type Settings struct {
	WindowWidth      int    `json:"window_width"`
	WindowHeight     int    `json:"window_height"`
	WindowMaximized  bool   `json:"window_maximized"`
	Theme            string `json:"theme"`
	MinimapThreshold int    `json:"minimap_threshold"`
}

var CurrentSettings *Settings

func DefaultSettings() *Settings {
	return &Settings{
		WindowWidth:      1920,
		WindowHeight:     1080,
		WindowMaximized:  false,
		Theme:            "Dark",
		MinimapThreshold: 10,
	}
}

func GetSettingsPath() string {
	// For now, save next to executable or in current dir
	return "settings.json"
}

func LoadSettings() *Settings {
	path := GetSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings()
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}

	// Validate basic sanity
	if s.WindowWidth < 100 {
		s.WindowWidth = 800
	}
	if s.WindowHeight < 100 {
		s.WindowHeight = 600
	}

	return &s
}

func SaveSettings(s *Settings) error {
	path := GetSettingsPath()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

