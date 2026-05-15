package model

type Keycode struct {
	Code  uint16 `json:"code"`
	Hex   string `json:"hex"`
	Label string `json:"label"`
}

type State struct {
	Label     string `json:"label"`
	Top       int    `json:"top"`
	Name      string `json:"name"`
	Effective string `json:"effective"`
	Temp      string `json:"temp"`
	Default   string `json:"default"`
	Tooltip   string `json:"tooltip,omitempty"`
}

type FeatureSet struct {
	LayerState   bool `json:"layer_state"`
	Encoders     bool `json:"encoders"`
	Combos       bool `json:"combos"`
	TapDance     bool `json:"tap_dance"`
	Macros       bool `json:"macros"`
	KeyOverrides bool `json:"key_overrides"`
	AltRepeat    bool `json:"alt_repeat"`
	QMKSettings  bool `json:"qmk_settings"`
}

type Capabilities struct {
	HelperProtocolVersion uint8      `json:"helper_protocol_version"`
	LayerCount            uint8      `json:"layer_count"`
	EncoderCount          uint8      `json:"encoder_count"`
	TapDanceEntries       uint8      `json:"tap_dance_entries"`
	ComboEntries          uint8      `json:"combo_entries"`
	KeyOverrideEntries    uint8      `json:"key_override_entries"`
	AltRepeatEntries      uint8      `json:"alt_repeat_entries"`
	MacroCount            uint8      `json:"macro_count"`
	MacroBufferSize       uint16     `json:"macro_buffer_size"`
	FeatureFlags          uint16     `json:"feature_flags"`
	Features              FeatureSet `json:"features"`
	VialProtocolVersion   uint32     `json:"vial_protocol_version"`
}

type Encoder struct {
	ID  int     `json:"id"`
	CCW Keycode `json:"ccw"`
	CW  Keycode `json:"cw"`
}

type Combo struct {
	Index  int       `json:"index"`
	Inputs []Keycode `json:"inputs"`
	Output Keycode   `json:"output"`
	Active bool      `json:"active"`
}

type TapDance struct {
	Index             int     `json:"index"`
	OnTap             Keycode `json:"on_tap"`
	OnHold            Keycode `json:"on_hold"`
	OnDoubleTap       Keycode `json:"on_double_tap"`
	OnTapHold         Keycode `json:"on_tap_hold"`
	CustomTappingTerm uint16  `json:"custom_tapping_term"`
	Active            bool    `json:"active"`
}

type MacroEntry struct {
	Index  int     `json:"index"`
	RawHex string  `json:"raw_hex"`
	Bytes  []uint8 `json:"bytes"`
	Active bool    `json:"active"`
}

type Macros struct {
	Count      int          `json:"count"`
	BufferSize int          `json:"buffer_size"`
	BufferHex  string       `json:"buffer_hex"`
	Entries    []MacroEntry `json:"entries"`
}

type KeyOverride struct {
	Index           int     `json:"index"`
	Trigger         Keycode `json:"trigger"`
	Replacement     Keycode `json:"replacement"`
	LayersMask      uint16  `json:"layers_mask"`
	TriggerMods     uint8   `json:"trigger_mods"`
	NegativeModMask uint8   `json:"negative_mod_mask"`
	SuppressedMods  uint8   `json:"suppressed_mods"`
	Options         uint8   `json:"options"`
	Active          bool    `json:"active"`
}

type AltRepeat struct {
	Index       int     `json:"index"`
	Keycode     Keycode `json:"keycode"`
	AltKeycode  Keycode `json:"alt_keycode"`
	AllowedMods uint8   `json:"allowed_mods"`
	Options     uint8   `json:"options"`
	Active      bool    `json:"active"`
}

type QMKSettings struct {
	Supported    bool `json:"supported"`
	APIAvailable bool `json:"api_available"`
	Decoded      bool `json:"decoded"`
}

type Matrix struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

type Layout struct {
	Helper        Capabilities           `json:"helper"`
	LayerCount    int                    `json:"layer_count"`
	Matrix        Matrix                 `json:"matrix"`
	LayerNames    map[string]string      `json:"layer_names"`
	Layers        map[string][][]Keycode `json:"layers"`
	Encoders      map[string][]Encoder   `json:"encoders"`
	Combos        []Combo                `json:"combos"`
	TapDances     []TapDance             `json:"tap_dances"`
	Macros        Macros                 `json:"macros"`
	KeyOverrides  []KeyOverride          `json:"key_overrides"`
	AltRepeatKeys []AltRepeat            `json:"alt_repeat_keys"`
	QMKSettings   QMKSettings            `json:"qmk_settings"`
}
