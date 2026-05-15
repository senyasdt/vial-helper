package keycodes

import (
	"fmt"

	"github.com/example/vial-helper/internal/model"
)

var basic = buildBasic()

func Payload(code uint16) model.Keycode {
	return model.Keycode{
		Code:  code,
		Hex:   fmt.Sprintf("0x%04X", code),
		Label: Label(code),
	}
}

func Label(code uint16) string {
	if label, ok := basic[code]; ok {
		return label
	}

	switch {
	case code >= 0x4000 && code <= 0x4FFF:
		layer := (code >> 8) & 0x000F
		tapped := code & 0x00FF
		return fmt.Sprintf("LT(%d,%s)", layer, basicLabel(tapped))
	case code >= 0x5200 && code <= 0x521F:
		return fmt.Sprintf("TO(%d)", code&0x001F)
	case code >= 0x5220 && code <= 0x523F:
		return fmt.Sprintf("MO(%d)", code&0x001F)
	case code >= 0x5240 && code <= 0x525F:
		return fmt.Sprintf("DF(%d)", code&0x001F)
	case code >= 0x5260 && code <= 0x527F:
		return fmt.Sprintf("TG(%d)", code&0x001F)
	case code >= 0x5280 && code <= 0x529F:
		return fmt.Sprintf("OSL(%d)", code&0x001F)
	case code >= 0x52C0 && code <= 0x52DF:
		return fmt.Sprintf("TT(%d)", code&0x001F)
	case code >= 0x52E0 && code <= 0x52FF:
		return fmt.Sprintf("PDF(%d)", code&0x001F)
	case code >= 0x5700 && code <= 0x57FF:
		return fmt.Sprintf("TD(%d)", code&0x00FF)
	case code >= 0x2000 && code <= 0x3FFF:
		tapped := code & 0x00FF
		return fmt.Sprintf("MT(…,%s)", basicLabel(tapped))
	}

	return fmt.Sprintf("0x%04X", code)
}

func basicLabel(code uint16) string {
	if label, ok := basic[code]; ok {
		return label
	}
	return fmt.Sprintf("KC_%02X", code)
}

func buildBasic() map[uint16]string {
	m := map[uint16]string{
		0x0000: "NO",
		0x0001: "TRNS",
		0x0028: "Enter",
		0x0029: "Esc",
		0x002A: "Backsp",
		0x002B: "Tab",
		0x002C: "Space",
		0x002D: "-",
		0x002E: "=",
		0x002F: "[",
		0x0030: "]",
		0x0031: "\\",
		0x0033: ";",
		0x0034: "'",
		0x0035: "`",
		0x0036: ",",
		0x0037: ".",
		0x0038: "/",
		0x0049: "Ins",
		0x004A: "Home",
		0x004B: "PgUp",
		0x004C: "Del",
		0x004D: "End",
		0x004E: "PgDn",
		0x004F: "→",
		0x0050: "←",
		0x0051: "↓",
		0x0052: "↑",
		0x00E0: "LCtrl",
		0x00E1: "LShift",
		0x00E2: "LAlt",
		0x00E3: "LWin",
		0x00E4: "RCtrl",
		0x00E5: "RShift",
		0x00E6: "RAlt",
		0x00E7: "RWin",
	}
	for i := 0; i < 26; i++ {
		m[0x0004+uint16(i)] = string(rune('A' + i))
	}
	digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	for i, digit := range digits {
		m[0x001E+uint16(i)] = digit
	}
	for i := 0; i < 12; i++ {
		m[0x003A+uint16(i)] = fmt.Sprintf("F%d", i+1)
	}
	for i := 12; i < 24; i++ {
		m[0x0068+uint16(i-12)] = fmt.Sprintf("F%d", i+1)
	}
	return m
}
