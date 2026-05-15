package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/example/vial-helper/internal/config"
	"github.com/example/vial-helper/internal/hidclient"
	"github.com/example/vial-helper/internal/model"
	"github.com/example/vial-helper/internal/protocol"
)

type DoctorOptions struct {
	Raw bool
}

type DoctorReport struct {
	Healthy         bool
	ConfigFile      string
	StateFile       string
	LayoutFile      string
	RefreshFlag     string
	Capabilities    model.Capabilities
	State           model.State
	Layout          model.Layout
	ActiveCombos    int
	ActiveTapDance  int
	ActiveMacros    int
	ActiveOverrides int
	ActiveAltRepeat int
	EncoderSummary  []string
	Warnings        []string
	Errors          []string
	CheckedAt       time.Time
}

func Doctor(cfg config.Config, paths config.Paths, opts DoctorOptions, out io.Writer) error {
	report := DoctorReport{
		ConfigFile:  paths.ConfigFile,
		StateFile:   paths.StateFile,
		LayoutFile:  paths.LayoutFile,
		RefreshFlag: paths.RefreshFlag,
		CheckedAt:   time.Now(),
	}

	client, err := hidclient.Open(
		cfg,
		time.Duration(cfg.Polling.RequestTimeoutMS)*time.Millisecond,
		protocol.Probe,
	)
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("open helper HID device: %v", err))
		writeDoctorReport(out, report, opts)
		return err
	}
	defer client.Close()

	reader := protocol.NewReader(client, cfg)

	caps, err := reader.Capabilities()
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("read capabilities: %v", err))
		writeDoctorReport(out, report, opts)
		return err
	}
	report.Capabilities = caps

	if caps.HelperProtocolVersion != protocol.HelperProtocolVersion {
		report.Warnings = append(report.Warnings, fmt.Sprintf(
			"helper protocol version mismatch: firmware=%d expected=%d",
			caps.HelperProtocolVersion,
			protocol.HelperProtocolVersion,
		))
	}

	state, err := reader.LayerState(cfg.Layers)
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("read layer state: %v", err))
		writeDoctorReport(out, report, opts)
		return err
	}
	report.State = state

	layout, err := reader.Layout()
	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("read full layout snapshot: %v", err))
		writeDoctorReport(out, report, opts)
		return err
	}
	report.Layout = layout

	report.ActiveCombos = countActiveCombos(layout.Combos)
	report.ActiveTapDance = countActiveTapDances(layout.TapDances)
	report.ActiveMacros = countActiveMacros(layout.Macros.Entries)
	report.ActiveOverrides = countActiveOverrides(layout.KeyOverrides)
	report.ActiveAltRepeat = countActiveAltRepeat(layout.AltRepeatKeys)
	report.EncoderSummary = summarizeEncoders(layout.Encoders["0"])

	report.Healthy = len(report.Errors) == 0
	writeDoctorReport(out, report, opts)
	return nil
}

func writeDoctorReport(out io.Writer, report DoctorReport, opts DoctorOptions) {
	fmt.Fprintln(out, "Vial Helper Doctor")
	fmt.Fprintln(out, strings.Repeat("=", 72))
	fmt.Fprintf(out, "Checked at: %s\n\n", report.CheckedAt.Format(time.RFC3339))

	fmt.Fprintln(out, "Paths")
	fmt.Fprintf(out, "  Config:       %s\n", report.ConfigFile)
	fmt.Fprintf(out, "  State:        %s\n", report.StateFile)
	fmt.Fprintf(out, "  Layout:       %s\n", report.LayoutFile)
	fmt.Fprintf(out, "  Refresh flag: %s\n\n", report.RefreshFlag)

	if len(report.Errors) > 0 {
		fmt.Fprintln(out, "Errors")
		for _, msg := range report.Errors {
			fmt.Fprintf(out, "  - %s\n", msg)
		}
		fmt.Fprintln(out)
	}

	if len(report.Warnings) > 0 {
		fmt.Fprintln(out, "Warnings")
		for _, msg := range report.Warnings {
			fmt.Fprintf(out, "  - %s\n", msg)
		}
		fmt.Fprintln(out)
	}

	if report.Capabilities.LayerCount > 0 {
		c := report.Capabilities
		fmt.Fprintln(out, "Capabilities")
		fmt.Fprintf(out, "  Helper protocol:    %d\n", c.HelperProtocolVersion)
		fmt.Fprintf(out, "  Vial protocol:      %d\n", c.VialProtocolVersion)
		fmt.Fprintf(out, "  Feature flags:      0x%04X\n", c.FeatureFlags)
		fmt.Fprintf(out, "  Layers:             %d\n", c.LayerCount)
		fmt.Fprintf(out, "  Encoders:           %d\n", c.EncoderCount)
		fmt.Fprintf(out, "  Combos:             %d\n", c.ComboEntries)
		fmt.Fprintf(out, "  Tap Dance:          %d\n", c.TapDanceEntries)
		fmt.Fprintf(out, "  Macros:             %d\n", c.MacroCount)
		fmt.Fprintf(out, "  Macro buffer:       %d bytes\n", c.MacroBufferSize)
		fmt.Fprintf(out, "  Key Overrides:      %d\n", c.KeyOverrideEntries)
		fmt.Fprintf(out, "  Alt Repeat:         %d\n", c.AltRepeatEntries)
		fmt.Fprintf(out, "  QMK Settings:       %t\n\n", c.Features.QMKSettings)
	}

	if report.State.Name != "" {
		s := report.State
		fmt.Fprintln(out, "Layer state")
		fmt.Fprintf(out, "  Current layer:      %s (%d)\n", s.Name, s.Top)
		fmt.Fprintf(out, "  Effective:          %s\n", s.Effective)
		fmt.Fprintf(out, "  Temporary:          %s\n", s.Temp)
		fmt.Fprintf(out, "  Default:            %s\n\n", s.Default)
	}

	if report.Layout.LayerCount > 0 {
		fmt.Fprintln(out, "Snapshot checks")
		fmt.Fprintf(out, "  Layers decoded:     %d\n", report.Layout.LayerCount)
		fmt.Fprintf(out, "  Active combos:      %d / %d\n", report.ActiveCombos, len(report.Layout.Combos))
		fmt.Fprintf(out, "  Active Tap Dance:   %d / %d\n", report.ActiveTapDance, len(report.Layout.TapDances))
		fmt.Fprintf(out, "  Active macros:      %d / %d\n", report.ActiveMacros, len(report.Layout.Macros.Entries))
		fmt.Fprintf(out, "  Active overrides:   %d / %d\n", report.ActiveOverrides, len(report.Layout.KeyOverrides))
		fmt.Fprintf(out, "  Active Alt Repeat:  %d / %d\n", report.ActiveAltRepeat, len(report.Layout.AltRepeatKeys))
		if len(report.EncoderSummary) > 0 {
			fmt.Fprintln(out, "  Base layer encoders:")
			for _, line := range report.EncoderSummary {
				fmt.Fprintf(out, "    %s\n", line)
			}
		}
		fmt.Fprintln(out)
	}

	if opts.Raw && report.Layout.LayerCount > 0 {
		fmt.Fprintln(out, "Raw JSON snapshot")
		fmt.Fprintln(out, strings.Repeat("-", 72))
		raw, err := json.MarshalIndent(report.Layout, "", "  ")
		if err != nil {
			fmt.Fprintf(out, "  failed to encode layout JSON: %v\n", err)
		} else {
			fmt.Fprintln(out, string(raw))
		}
		fmt.Fprintln(out)
	}

	result := "HEALTHY"
	if !report.Healthy {
		result = "UNHEALTHY"
	}
	fmt.Fprintf(out, "Result: %s\n", result)
}

func countActiveCombos(items []model.Combo) int {
	n := 0
	for _, item := range items {
		if item.Active {
			n++
		}
	}
	return n
}

func countActiveTapDances(items []model.TapDance) int {
	n := 0
	for _, item := range items {
		if item.Active {
			n++
		}
	}
	return n
}

func countActiveMacros(items []model.MacroEntry) int {
	n := 0
	for _, item := range items {
		if item.Active {
			n++
		}
	}
	return n
}

func countActiveOverrides(items []model.KeyOverride) int {
	n := 0
	for _, item := range items {
		if item.Active {
			n++
		}
	}
	return n
}

func countActiveAltRepeat(items []model.AltRepeat) int {
	n := 0
	for _, item := range items {
		if item.Active {
			n++
		}
	}
	return n
}

func summarizeEncoders(items []model.Encoder) []string {
	if len(items) == 0 {
		return nil
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprintf("encoder %d: %s / %s", item.ID, item.CCW.Label, item.CW.Label))
	}
	return out
}

type StatusReport struct {
	ConfigFile    string
	StateFile     string
	LayoutFile    string
	RefreshFlag   string
	StateExists   bool
	LayoutExists  bool
	StateModTime  time.Time
	LayoutModTime time.Time
	State         *model.State
	LayoutHelper  *model.Capabilities
}

func Status(paths config.Paths, out io.Writer) error {
	report := StatusReport{
		ConfigFile:  paths.ConfigFile,
		StateFile:   paths.StateFile,
		LayoutFile:  paths.LayoutFile,
		RefreshFlag: paths.RefreshFlag,
	}

	if info, err := os.Stat(paths.StateFile); err == nil {
		report.StateExists = true
		report.StateModTime = info.ModTime()
		if raw, readErr := os.ReadFile(paths.StateFile); readErr == nil {
			var state model.State
			if jsonErr := json.Unmarshal(raw, &state); jsonErr == nil {
				report.State = &state
			}
		}
	}

	if info, err := os.Stat(paths.LayoutFile); err == nil {
		report.LayoutExists = true
		report.LayoutModTime = info.ModTime()
		if raw, readErr := os.ReadFile(paths.LayoutFile); readErr == nil {
			var layout struct {
				Helper model.Capabilities `json:"helper"`
			}
			if jsonErr := json.Unmarshal(raw, &layout); jsonErr == nil {
				report.LayoutHelper = &layout.Helper
			}
		}
	}

	fmt.Fprintln(out, "Vial Helper Status")
	fmt.Fprintln(out, strings.Repeat("=", 72))
	fmt.Fprintf(out, "Config:       %s\n", report.ConfigFile)
	fmt.Fprintf(out, "State:        %s\n", report.StateFile)
	fmt.Fprintf(out, "Layout:       %s\n", report.LayoutFile)
	fmt.Fprintf(out, "Refresh flag: %s\n\n", report.RefreshFlag)

	fmt.Fprintf(out, "state.json:   %s\n", statusFileLabel(report.StateExists, report.StateModTime))
	if report.State != nil {
		fmt.Fprintf(out, "  layer:      %s (%d)\n", report.State.Name, report.State.Top)
		fmt.Fprintf(out, "  effective:  %s\n", report.State.Effective)
	}
	fmt.Fprintf(out, "layout.json:  %s\n", statusFileLabel(report.LayoutExists, report.LayoutModTime))
	if report.LayoutHelper != nil {
		fmt.Fprintf(out, "  protocol:   helper=%d vial=%d\n", report.LayoutHelper.HelperProtocolVersion, report.LayoutHelper.VialProtocolVersion)
		fmt.Fprintf(out, "  layers:     %d\n", report.LayoutHelper.LayerCount)
		fmt.Fprintf(out, "  encoders:   %d\n", report.LayoutHelper.EncoderCount)
	}
	return nil
}

func statusFileLabel(exists bool, modTime time.Time) string {
	if !exists {
		return "missing"
	}
	return fmt.Sprintf("present, updated %s", modTime.Format(time.RFC3339))
}
