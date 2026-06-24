package main

import (
	"encoding/json"
	"os"
)

type DelayPreset string

const (
	PresetSlow   DelayPreset = "slow"
	PresetNormal DelayPreset = "normal"
	PresetFast   DelayPreset = "fast"
	PresetCustom DelayPreset = "custom"
)

type Config struct {
	GASURL         string      `json:":"`
	DelayPreset    DelayPreset `json:"delay_preset"`
	DelayCustomMin int         `json:"delay_custom_min"`
	DelayCustomMax int         `json:"delay_custom_max"`
	// Human-like typing settings
	HumanLikeTyping bool `json:"human_like_typing"`
	BurstLengthMin  int  `json:"burst_length_min"`
	BurstLengthMax  int  `json:"burst_length_max"`
	LongPauseFreq   int  `json:"long_pause_freq"`
	LongPauseMin    int  `json:"long_pause_min"`
	LongPauseMax    int  `json:"long_pause_max"`
	ErrorRate       int  `json:"error_rate"`
	CorrectionRate  int  `json:"correction_rate"`
	TypoRate        int  `json:"typo_rate"` // 0–100, 0 = tidak ada typo
}

var defaultConfig = Config{
	GASURL:          "YOUR_GAS_URL_HERE",  // Replace with your Google Apps Script URL
	DelayPreset:     PresetNormal,
	DelayCustomMin:  60,
	DelayCustomMax:  130,
	HumanLikeTyping: true, // Enable human-like typing by default
	BurstLengthMin:  3,
	BurstLengthMax:  10,
	LongPauseFreq:   50, // On average, every 50 characters
	LongPauseMin:    500,
	LongPauseMax:    1500,
	ErrorRate:       2,  // 2% chance of typo
	CorrectionRate:  80, // 80% chance of correcting a typo
	TypoRate:        0,
}

// MinMax mengembalikan range delay (ms) sesuai preset aktif
func (c Config) MinMax() (min, max int) {
	switch c.DelayPreset {
	case PresetSlow:
		return 80, 160
	case PresetFast:
		return 30, 80
	case PresetCustom:
		return c.DelayCustomMin, c.DelayCustomMax
	default: // normal
		return 60, 130
	}
}

const configPath = "config.json"

func loadConfig() (Config, error) {
	cfg := defaultConfig
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Buat file baru dengan default
		return cfg, saveConfig(cfg)
	}
	if err != nil {
		return cfg, err
	}
	return cfg, json.Unmarshal(data, &cfg)
}

func saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}