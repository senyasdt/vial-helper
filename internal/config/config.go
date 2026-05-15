package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const AppName = "vial-helper"

type Device struct {
	VID       uint16 `toml:"vid"`
	PID       uint16 `toml:"pid"`
	UsagePage uint16 `toml:"usage_page"`
	Usage     uint16 `toml:"usage"`
}

type Polling struct {
	LayerPollMS      int `toml:"layer_poll_ms"`
	ReconnectMS      int `toml:"reconnect_ms"`
	RequestTimeoutMS int `toml:"request_timeout_ms"`
}

type Output struct {
	Dir         string `toml:"dir"`
	StateFile   string `toml:"state_file"`
	LayoutFile  string `toml:"layout_file"`
	RefreshFlag string `toml:"refresh_flag"`
}

type Config struct {
	Device  Device            `toml:"device"`
	Polling Polling           `toml:"polling"`
	Output  Output            `toml:"output"`
	Layers  map[string]string `toml:"layers"`
}

type Paths struct {
	ConfigDir   string
	ConfigFile  string
	OutputDir   string
	StateFile   string
	LayoutFile  string
	RefreshFlag string
}

func Default() Config {
	layers := make(map[string]string, 16)
	layers["0"] = "BASE"
	for i := 1; i < 16; i++ {
		layers[strconv.Itoa(i)] = fmt.Sprintf("L%d", i)
	}

	return Config{
		Device: Device{
			VID:       0xFEED,
			PID:       0x4079,
			UsagePage: 0xFF60,
			Usage:     0x0061,
		},
		Polling: Polling{
			LayerPollMS:      50,
			ReconnectMS:      1000,
			RequestTimeoutMS: 500,
		},
		Output: Output{
			Dir:         "",
			StateFile:   "state.json",
			LayoutFile:  "layout.json",
			RefreshFlag: "refresh-layout.flag",
		},
		Layers: layers,
	}
}

func AppConfigDir() (string, error) {
	if raw := strings.TrimSpace(os.Getenv("VIAL_HELPER_CONFIG_DIR")); raw != "" {
		return expandHome(raw)
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, AppName), nil
}

func Resolve(configPath string) (Config, Paths, error) {
	configDir, err := AppConfigDir()
	if err != nil {
		return Config{}, Paths{}, err
	}

	if strings.TrimSpace(configPath) == "" {
		configPath = filepath.Join(configDir, "config.toml")
	} else {
		configPath, err = expandHome(configPath)
		if err != nil {
			return Config{}, Paths{}, err
		}
		configDir = filepath.Dir(configPath)
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return Config{}, Paths{}, fmt.Errorf("create config dir: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if writeErr := WriteDefault(configPath); writeErr != nil {
			return Config{}, Paths{}, writeErr
		}
	} else if err != nil {
		return Config{}, Paths{}, fmt.Errorf("stat config file: %w", err)
	}

	cfg := Default()
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, Paths{}, fmt.Errorf("decode config: %w", err)
	}
	cfg.Normalize()

	outputDir := strings.TrimSpace(cfg.Output.Dir)
	if outputDir == "" {
		outputDir = configDir
	} else {
		outputDir, err = expandHome(outputDir)
		if err != nil {
			return Config{}, Paths{}, err
		}
		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(configDir, outputDir)
		}
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Config{}, Paths{}, fmt.Errorf("create output dir: %w", err)
	}

	paths := Paths{
		ConfigDir:   configDir,
		ConfigFile:  configPath,
		OutputDir:   outputDir,
		StateFile:   resolveOutputPath(outputDir, cfg.Output.StateFile, "state.json"),
		LayoutFile:  resolveOutputPath(outputDir, cfg.Output.LayoutFile, "layout.json"),
		RefreshFlag: resolveOutputPath(outputDir, cfg.Output.RefreshFlag, "refresh-layout.flag"),
	}
	return cfg, paths, nil
}

func (c *Config) Normalize() {
	if c.Device.VID == 0 {
		c.Device.VID = 0xFEED
	}
	if c.Device.PID == 0 {
		c.Device.PID = 0x4079
	}
	if c.Device.UsagePage == 0 {
		c.Device.UsagePage = 0xFF60
	}
	if c.Device.Usage == 0 {
		c.Device.Usage = 0x0061
	}
	if c.Polling.LayerPollMS <= 0 {
		c.Polling.LayerPollMS = 50
	}
	if c.Polling.ReconnectMS <= 0 {
		c.Polling.ReconnectMS = 1000
	}
	if c.Polling.RequestTimeoutMS <= 0 {
		c.Polling.RequestTimeoutMS = 500
	}
	if strings.TrimSpace(c.Output.StateFile) == "" {
		c.Output.StateFile = "state.json"
	}
	if strings.TrimSpace(c.Output.LayoutFile) == "" {
		c.Output.LayoutFile = "layout.json"
	}
	if strings.TrimSpace(c.Output.RefreshFlag) == "" {
		c.Output.RefreshFlag = "refresh-layout.flag"
	}
	if c.Layers == nil {
		c.Layers = Default().Layers
	}
}

func WriteDefault(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config parent dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(DefaultConfigTOML()), 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}
	return nil
}

func DefaultConfigTOML() string {
	return `# vial-helper configuration

[device]
vid = 0xFEED
pid = 0x4079
usage_page = 0xFF60
usage = 0x0061

[polling]
layer_poll_ms = 50
reconnect_ms = 1000
request_timeout_ms = 500

[output]
# Empty means the same directory as config.toml.
dir = ""
state_file = "state.json"
layout_file = "layout.json"
refresh_flag = "refresh-layout.flag"

[layers]
0 = "BASE"
1 = "L1"
2 = "L2"
3 = "L3"
4 = "L4"
5 = "L5"
6 = "L6"
7 = "L7"
8 = "L8"
9 = "L9"
10 = "L10"
11 = "L11"
12 = "L12"
13 = "L13"
14 = "L14"
15 = "L15"
`
}

func resolveOutputPath(outputDir, configured, fallback string) string {
	value := strings.TrimSpace(configured)
	if value == "" {
		value = fallback
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(outputDir, value)
}

func expandHome(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", nil
	}
	if trimmed == "~" || strings.HasPrefix(trimmed, "~/") || strings.HasPrefix(trimmed, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		if trimmed == "~" {
			return home, nil
		}
		return filepath.Join(home, trimmed[2:]), nil
	}
	return trimmed, nil
}
