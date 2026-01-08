package engine

// ToneRule defines which tone placement rule to use
type ToneRule int

const (
	// ToneRuleOld is the traditional rule (quy tắc cũ)
	// - hoà (on 'o'), của (on 'u'), mùa (on 'u')
	ToneRuleOld ToneRule = iota

	// ToneRuleNew is the modern rule (quy tắc mới)
	// - hòa (on 'a'), của (on 'a'), mùa (on 'a')
	ToneRuleNew
)

// EngineConfig holds configuration options for the engine
type EngineConfig struct {
	// ToneRule determines which tone placement rule to use
	ToneRule ToneRule

	// EnableValidation enables Vietnamese validation before transformation
	// When true, non-Vietnamese text won't be transformed
	EnableValidation bool

	// EnableDoubleKeyRevert allows reverting transformations by pressing key again
	// e.g., "aa" -> "â" -> "aa" (on third 'a')
	EnableDoubleKeyRevert bool

	// EnableWAsVowel allows single 'w' to become 'ư' when valid
	EnableWAsVowel bool

	// InputMethodName specifies which input method to use ("Telex" or "VNI")
	InputMethodName string
}

// DefaultConfig returns the default engine configuration
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		ToneRule:              ToneRuleOld, // Traditional rule by default
		EnableValidation:      true,        // Enable validation
		EnableDoubleKeyRevert: true,        // Enable double-key revert
		EnableWAsVowel:        true,        // Enable W as vowel
		InputMethodName:       "Telex",     // Default to Telex
	}
}

// ConfiguredEngine is an extended composition engine with configuration
type ConfiguredEngine struct {
	*CompositionEngine
	config *EngineConfig
}

// NewConfiguredEngine creates an engine with the given configuration
func NewConfiguredEngine(config *EngineConfig) *ConfiguredEngine {
	if config == nil {
		config = DefaultConfig()
	}

	engine := NewCompositionEngine()

	// Set input method based on config
	switch config.InputMethodName {
	case "VNI":
		engine.SetInputMethod(NewVNIMethod())
	default:
		engine.SetInputMethod(NewTelexMethod())
	}

	return &ConfiguredEngine{
		CompositionEngine: engine,
		config:            config,
	}
}

// SetConfig updates the engine configuration
func (e *ConfiguredEngine) SetConfig(config *EngineConfig) {
	e.config = config

	// Update input method if changed
	switch config.InputMethodName {
	case "VNI":
		e.SetInputMethod(NewVNIMethod())
	default:
		e.SetInputMethod(NewTelexMethod())
	}
}

// GetConfig returns the current configuration
func (e *ConfiguredEngine) GetConfig() *EngineConfig {
	return e.config
}

// SetToneRule sets the tone placement rule
func (e *ConfiguredEngine) SetToneRule(rule ToneRule) {
	e.config.ToneRule = rule
}

// SetEnableValidation enables or disables Vietnamese validation
func (e *ConfiguredEngine) SetEnableValidation(enable bool) {
	e.config.EnableValidation = enable
}

// SetEnableDoubleKeyRevert enables or disables double-key revert
func (e *ConfiguredEngine) SetEnableDoubleKeyRevert(enable bool) {
	e.config.EnableDoubleKeyRevert = enable
}

// SetEnableWAsVowel enables or disables W-as-vowel feature
func (e *ConfiguredEngine) SetEnableWAsVowel(enable bool) {
	e.config.EnableWAsVowel = enable
}

// UsesModernToneRule returns true if using the modern tone placement rule
func (e *ConfiguredEngine) UsesModernToneRule() bool {
	return e.config.ToneRule == ToneRuleNew
}
