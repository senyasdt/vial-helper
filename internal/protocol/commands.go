package protocol

const (
	CmdCapabilities  byte = 0x40
	CmdLayerState    byte = 0x42
	CmdEncoderAction byte = 0x43
	CmdComboEntry    byte = 0x44
	CmdTapDanceEntry byte = 0x45
	CmdMacro         byte = 0x46
	CmdKeyOverride   byte = 0x47
	CmdAltRepeat     byte = 0x48
	CmdQMKSetting    byte = 0x49

	HelperProtocolVersion byte = 0x02

	MacroInfo  byte = 0x00
	MacroChunk byte = 0x01

	VIAGetKeycode byte = 0x04

	FeatureLayerState  uint16 = 1 << 0
	FeatureEncoders    uint16 = 1 << 1
	FeatureCombos      uint16 = 1 << 2
	FeatureTapDance    uint16 = 1 << 3
	FeatureMacros      uint16 = 1 << 4
	FeatureOverrides   uint16 = 1 << 5
	FeatureAltRepeat   uint16 = 1 << 6
	FeatureQMKSettings uint16 = 1 << 7

	MatrixRows = 4
	MatrixCols = 4
)
