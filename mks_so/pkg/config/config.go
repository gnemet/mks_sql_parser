package config

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/spf13/viper"
)

type PatternConfig struct {
	ID          int    `mapstructure:"id" json:"id"`
	Regex       string `mapstructure:"regex" json:"regex"`
	Action      string `mapstructure:"action" json:"action"`
	Description string `mapstructure:"description" json:"description"`
}

type CompiledPattern struct {
	Re     *regexp.Regexp
	Action string
}

func LoadPatterns() []CompiledPattern {
	configs, err := LoadConfigs(nil)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	var compiledPatterns []CompiledPattern
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			fmt.Printf("Error compiling regex '%s': %s\n", cfg.Regex, err)
			continue
		}
		compiledPatterns = append(compiledPatterns, CompiledPattern{Re: re, Action: cfg.Action})
	}
	return compiledPatterns
}

// LoadPatternsWithConfig allows loading from specific bytes (e.g. embedded)
func LoadPatternsWithConfig(configBytes []byte) []CompiledPattern {
	configs, err := LoadConfigs(configBytes)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	var compiledPatterns []CompiledPattern
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			fmt.Printf("Error compiling regex '%s': %s\n", cfg.Regex, err)
			continue
		}
		compiledPatterns = append(compiledPatterns, CompiledPattern{Re: re, Action: cfg.Action})
	}
	return compiledPatterns
}

func LoadConfigs(configBytes []byte) ([]PatternConfig, error) {
	viper.Reset() // Clear previous config
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return nil, fmt.Errorf("error reading config bytes: %w", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		viper.AddConfigPath("..")
		// Also add module root if running from pkg/config
		viper.AddConfigPath("../../..")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("config.yaml not found: %w", err)
		}
	}

	var configs []PatternConfig
	err := viper.UnmarshalKey("patterns", &configs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling patterns: %w", err)
	}
	return configs, nil
}

type TestCase struct {
	ID       int                    `mapstructure:"id" json:"id"`
	Input    map[string]interface{} `mapstructure:"input" json:"input"`
	Text     string                 `mapstructure:"text" json:"text"`
	Expected string                 `mapstructure:"expected" json:"expected"`
	Passed   bool                   `mapstructure:"passed" json:"passed"`
}

func LoadTests() []TestCase {
	// Re-read config in case it changed or wasn't loaded (though usually LoadPatterns is called first)
	// For safety, we can rely on init or just ensure viper has read it.
	// Assuming LoadPatterns or init has set up viper paths.
	// If independent, we might need to re-add paths, but let's assume viper is singleton and configured.
	// However, to be safe for a test runner that might only call LoadTests:
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.AddConfigPath("..")
	// Also add module root if running from pkg/config
	viper.AddConfigPath("../../..")

	if err := viper.ReadInConfig(); err != nil {
		// If it fails, maybe it's already loaded? Or file missing.
		// For tests, we want to know if it fails.
		fmt.Printf("Error reading config for tests: %v\n", err)
	}

	var tests []TestCase
	err := viper.UnmarshalKey("test", &tests)
	if err != nil {
		fmt.Printf("Error unmarshalling tests: %v\n", err)
	}
	return tests
}
