package protocol

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/example/vial-helper/internal/config"
	"github.com/example/vial-helper/internal/hidclient"
	"github.com/example/vial-helper/internal/keycodes"
	"github.com/example/vial-helper/internal/model"
)

type Reader struct {
	client *hidclient.Client
	cfg    config.Config
}

func NewReader(client *hidclient.Client, cfg config.Config) *Reader {
	return &Reader{client: client, cfg: cfg}
}

func Probe(c *hidclient.Client) error {
	resp, err := c.Send([]byte{CmdCapabilities}, CmdCapabilities)
	if err != nil {
		return err
	}
	if len(resp) < 3 {
		return fmt.Errorf("capabilities response too short: %d", len(resp))
	}
	if resp[1] != HelperProtocolVersion || resp[2] == 0 {
		return fmt.Errorf("unsupported helper protocol: version=%d layers=%d", resp[1], resp[2])
	}
	return nil
}

func (r *Reader) Capabilities() (model.Capabilities, error) {
	resp, err := r.client.Send([]byte{CmdCapabilities}, CmdCapabilities)
	if err != nil {
		return model.Capabilities{}, err
	}
	if len(resp) < 17 {
		return model.Capabilities{}, fmt.Errorf("capabilities response too short: %d", len(resp))
	}
	flags := readU16LE(resp, 11)
	caps := model.Capabilities{
		HelperProtocolVersion: resp[1],
		LayerCount:            resp[2],
		EncoderCount:          resp[3],
		TapDanceEntries:       resp[4],
		ComboEntries:          resp[5],
		KeyOverrideEntries:    resp[6],
		AltRepeatEntries:      resp[7],
		MacroCount:            resp[8],
		MacroBufferSize:       readU16LE(resp, 9),
		FeatureFlags:          flags,
		VialProtocolVersion:   readU32LE(resp, 13),
	}
	caps.Features = model.FeatureSet{
		LayerState:   flags&FeatureLayerState != 0,
		Encoders:     flags&FeatureEncoders != 0,
		Combos:       flags&FeatureCombos != 0,
		TapDance:     flags&FeatureTapDance != 0,
		Macros:       flags&FeatureMacros != 0,
		KeyOverrides: flags&FeatureOverrides != 0,
		AltRepeat:    flags&FeatureAltRepeat != 0,
		QMKSettings:  flags&FeatureQMKSettings != 0,
	}
	return caps, nil
}

func (r *Reader) LayerState(layerNames map[string]string) (model.State, error) {
	resp, err := r.client.Send([]byte{CmdLayerState}, CmdLayerState)
	if err != nil {
		return model.State{}, err
	}
	if len(resp) < 15 {
		return model.State{}, fmt.Errorf("layer state response too short: %d", len(resp))
	}
	top := int(resp[2])
	name := layerNames[fmt.Sprintf("%d", top)]
	if name == "" {
		name = fmt.Sprintf("L%d", top)
	}
	effective := maskToLayers(readU32LE(resp, 3), 16)
	temp := maskToLayers(readU32LE(resp, 7), 16)
	def := maskToLayers(readU32LE(resp, 11), 16)
	return model.State{
		Label:     name,
		Top:       top,
		Name:      name,
		Effective: formatLayers(effective),
		Temp:      formatLayers(temp),
		Default:   formatLayers(def),
		Tooltip: fmt.Sprintf(
			"Top: %s\nEffective: %s\nTemporary: %s\nDefault: %s",
			name,
			formatLayers(effective),
			formatLayers(temp),
			formatLayers(def),
		),
	}, nil
}

func (r *Reader) Layout() (model.Layout, error) {
	caps, err := r.Capabilities()
	if err != nil {
		return model.Layout{}, err
	}
	layers, err := r.layers(int(caps.LayerCount))
	if err != nil {
		return model.Layout{}, err
	}
	encoders, err := r.encoders(int(caps.LayerCount), int(caps.EncoderCount))
	if err != nil {
		return model.Layout{}, err
	}
	combos, err := r.combos(int(caps.ComboEntries))
	if err != nil {
		return model.Layout{}, err
	}
	tapDances, err := r.tapDances(int(caps.TapDanceEntries))
	if err != nil {
		return model.Layout{}, err
	}
	macros, err := r.macros()
	if err != nil {
		return model.Layout{}, err
	}
	overrides, err := r.overrides(int(caps.KeyOverrideEntries))
	if err != nil {
		return model.Layout{}, err
	}
	altRepeats, err := r.altRepeats(int(caps.AltRepeatEntries))
	if err != nil {
		return model.Layout{}, err
	}

	return model.Layout{
		Helper:        caps,
		LayerCount:    int(caps.LayerCount),
		Matrix:        model.Matrix{Rows: MatrixRows, Cols: MatrixCols},
		LayerNames:    r.cfg.Layers,
		Layers:        layers,
		Encoders:      encoders,
		Combos:        combos,
		TapDances:     tapDances,
		Macros:        macros,
		KeyOverrides:  overrides,
		AltRepeatKeys: altRepeats,
		QMKSettings: model.QMKSettings{
			Supported:    caps.Features.QMKSettings,
			APIAvailable: true,
			Decoded:      false,
		},
	}, nil
}

func (r *Reader) layers(layerCount int) (map[string][][]model.Keycode, error) {
	out := make(map[string][][]model.Keycode, layerCount)
	for layer := 0; layer < layerCount; layer++ {
		rows := make([][]model.Keycode, 0, MatrixRows)
		for row := 0; row < MatrixRows; row++ {
			cols := make([]model.Keycode, 0, MatrixCols)
			for col := 0; col < MatrixCols; col++ {
				code, err := r.keycode(layer, row, col)
				if err != nil {
					return nil, err
				}
				cols = append(cols, keycodes.Payload(code))
			}
			rows = append(rows, cols)
		}
		out[fmt.Sprintf("%d", layer)] = rows
	}
	return out, nil
}

func (r *Reader) keycode(layer, row, col int) (uint16, error) {
	resp, err := r.client.Send([]byte{VIAGetKeycode, byte(layer), byte(row), byte(col)}, VIAGetKeycode)
	if err != nil {
		return 0, err
	}
	if len(resp) < 6 {
		return 0, fmt.Errorf("keycode response too short: %d", len(resp))
	}
	return uint16(resp[4])<<8 | uint16(resp[5]), nil
}

func (r *Reader) encoders(layerCount, count int) (map[string][]model.Encoder, error) {
	out := make(map[string][]model.Encoder, layerCount)
	for layer := 0; layer < layerCount; layer++ {
		entries := make([]model.Encoder, 0, count)
		for id := 0; id < count; id++ {
			ccw, err := r.encoderAction(layer, id, false)
			if err != nil {
				return nil, err
			}
			cw, err := r.encoderAction(layer, id, true)
			if err != nil {
				return nil, err
			}
			entries = append(entries, model.Encoder{ID: id, CCW: keycodes.Payload(ccw), CW: keycodes.Payload(cw)})
		}
		out[fmt.Sprintf("%d", layer)] = entries
	}
	return out, nil
}

func (r *Reader) encoderAction(layer, id int, clockwise bool) (uint16, error) {
	dir := byte(0)
	if clockwise {
		dir = 1
	}
	resp, err := r.client.Send([]byte{CmdEncoderAction, byte(layer), byte(id), dir}, CmdEncoderAction)
	if err != nil {
		return 0, err
	}
	if len(resp) < 6 {
		return 0, fmt.Errorf("encoder response too short: %d", len(resp))
	}
	return uint16(resp[4])<<8 | uint16(resp[5]), nil
}

func (r *Reader) combos(count int) ([]model.Combo, error) {
	out := make([]model.Combo, 0, count)
	for i := 0; i < count; i++ {
		resp, err := r.client.Send([]byte{CmdComboEntry, byte(i)}, CmdComboEntry)
		if err != nil {
			return nil, err
		}
		if len(resp) < 13 {
			return nil, fmt.Errorf("combo response too short: %d", len(resp))
		}
		if resp[2] != 0 {
			continue
		}
		codes := []uint16{readU16LE(resp, 3), readU16LE(resp, 5), readU16LE(resp, 7), readU16LE(resp, 9)}
		inputs := make([]model.Keycode, 0, 4)
		active := false
		for _, code := range codes {
			inputs = append(inputs, keycodes.Payload(code))
			if code != 0 && code != 1 {
				active = true
			}
		}
		output := readU16LE(resp, 11)
		if output != 0 && output != 1 {
			active = true
		}
		out = append(out, model.Combo{Index: i, Inputs: inputs, Output: keycodes.Payload(output), Active: active})
	}
	return out, nil
}

func (r *Reader) tapDances(count int) ([]model.TapDance, error) {
	out := make([]model.TapDance, 0, count)
	for i := 0; i < count; i++ {
		resp, err := r.client.Send([]byte{CmdTapDanceEntry, byte(i)}, CmdTapDanceEntry)
		if err != nil {
			return nil, err
		}
		if len(resp) < 13 {
			return nil, fmt.Errorf("tap dance response too short: %d", len(resp))
		}
		if resp[2] != 0 {
			continue
		}
		onTap := readU16LE(resp, 3)
		onHold := readU16LE(resp, 5)
		onDouble := readU16LE(resp, 7)
		onTapHold := readU16LE(resp, 9)
		active := false
		for _, code := range []uint16{onTap, onHold, onDouble, onTapHold} {
			if code != 0 && code != 1 {
				active = true
			}
		}
		out = append(out, model.TapDance{
			Index:             i,
			OnTap:             keycodes.Payload(onTap),
			OnHold:            keycodes.Payload(onHold),
			OnDoubleTap:       keycodes.Payload(onDouble),
			OnTapHold:         keycodes.Payload(onTapHold),
			CustomTappingTerm: readU16LE(resp, 11),
			Active:            active,
		})
	}
	return out, nil
}

func (r *Reader) macros() (model.Macros, error) {
	info, err := r.client.Send([]byte{CmdMacro, MacroInfo}, CmdMacro)
	if err != nil {
		return model.Macros{}, err
	}
	if len(info) < 5 {
		return model.Macros{}, fmt.Errorf("macro info response too short: %d", len(info))
	}
	count := int(info[2])
	bufferSize := int(readU16LE(info, 3))
	buffer, err := r.macroBuffer(bufferSize)
	if err != nil {
		return model.Macros{}, err
	}
	return model.Macros{
		Count:      count,
		BufferSize: bufferSize,
		BufferHex:  spacedHex(buffer),
		Entries:    splitMacroEntries(buffer, count),
	}, nil
}

func (r *Reader) macroBuffer(size int) ([]byte, error) {
	out := make([]byte, 0, size)
	for offset := 0; offset < size; {
		want := size - offset
		if want > 27 {
			want = 27
		}
		resp, err := r.client.Send([]byte{CmdMacro, MacroChunk, byte(offset & 0xFF), byte((offset >> 8) & 0xFF), byte(want)}, CmdMacro)
		if err != nil {
			return nil, err
		}
		if len(resp) < 5 {
			return nil, fmt.Errorf("macro chunk response too short: %d", len(resp))
		}
		returned := int(resp[4])
		if returned == 0 {
			break
		}
		if 5+returned > len(resp) {
			return nil, fmt.Errorf("macro chunk out of bounds: returned=%d response=%d", returned, len(resp))
		}
		out = append(out, resp[5:5+returned]...)
		offset += returned
	}
	return out, nil
}

func splitMacroEntries(buffer []byte, count int) []model.MacroEntry {
	entries := make([]model.MacroEntry, 0, count)
	current := make([]byte, 0)
	for _, b := range buffer {
		if b == 0 {
			copied := append([]byte(nil), current...)
			entries = append(entries, model.MacroEntry{Index: len(entries), RawHex: spacedHex(copied), Bytes: copied, Active: len(copied) > 0})
			current = current[:0]
			if len(entries) >= count {
				break
			}
			continue
		}
		current = append(current, b)
	}
	for len(entries) < count {
		entries = append(entries, model.MacroEntry{Index: len(entries), Bytes: []uint8{}, Active: false})
	}
	return entries
}

func (r *Reader) overrides(count int) ([]model.KeyOverride, error) {
	out := make([]model.KeyOverride, 0, count)
	for i := 0; i < count; i++ {
		resp, err := r.client.Send([]byte{CmdKeyOverride, byte(i)}, CmdKeyOverride)
		if err != nil {
			return nil, err
		}
		if len(resp) < 13 {
			return nil, fmt.Errorf("key override response too short: %d", len(resp))
		}
		if resp[2] != 0 {
			continue
		}
		trigger := readU16LE(resp, 3)
		replacement := readU16LE(resp, 5)
		out = append(out, model.KeyOverride{
			Index:           i,
			Trigger:         keycodes.Payload(trigger),
			Replacement:     keycodes.Payload(replacement),
			LayersMask:      readU16LE(resp, 7),
			TriggerMods:     resp[9],
			NegativeModMask: resp[10],
			SuppressedMods:  resp[11],
			Options:         resp[12],
			Active:          trigger != 0 || replacement != 0,
		})
	}
	return out, nil
}

func (r *Reader) altRepeats(count int) ([]model.AltRepeat, error) {
	out := make([]model.AltRepeat, 0, count)
	for i := 0; i < count; i++ {
		resp, err := r.client.Send([]byte{CmdAltRepeat, byte(i)}, CmdAltRepeat)
		if err != nil {
			return nil, err
		}
		if len(resp) < 9 {
			return nil, fmt.Errorf("alt repeat response too short: %d", len(resp))
		}
		if resp[2] != 0 {
			continue
		}
		keycode := readU16LE(resp, 3)
		alt := readU16LE(resp, 5)
		out = append(out, model.AltRepeat{
			Index:       i,
			Keycode:     keycodes.Payload(keycode),
			AltKeycode:  keycodes.Payload(alt),
			AllowedMods: resp[7],
			Options:     resp[8],
			Active:      keycode != 0 || alt != 0,
		})
	}
	return out, nil
}

func readU16LE(data []byte, offset int) uint16 {
	if offset+1 >= len(data) {
		return 0
	}
	return uint16(data[offset]) | uint16(data[offset+1])<<8
}

func readU32LE(data []byte, offset int) uint32 {
	if offset+3 >= len(data) {
		return 0
	}
	return uint32(data[offset]) | uint32(data[offset+1])<<8 | uint32(data[offset+2])<<16 | uint32(data[offset+3])<<24
}

func maskToLayers(mask uint32, max int) []int {
	out := make([]int, 0)
	for i := 0; i < max; i++ {
		if mask&(1<<i) != 0 {
			out = append(out, i)
		}
	}
	return out
}

func formatLayers(layers []int) string {
	if len(layers) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(layers))
	for _, layer := range layers {
		parts = append(parts, fmt.Sprintf("%d", layer))
	}
	return strings.Join(parts, ", ")
}

func spacedHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	raw := strings.ToUpper(hex.EncodeToString(data))
	parts := make([]string, 0, len(raw)/2)
	for i := 0; i < len(raw); i += 2 {
		parts = append(parts, raw[i:i+2])
	}
	return strings.Join(parts, " ")
}
