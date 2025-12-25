package config

import (
	"fmt"
	"regexp"

	"github.com/spf13/viper"
)

type PatternConfig struct {
	Regex  string `mapstructure:"regex"`
	Action string `mapstructure:"action"`
}

type CompiledPattern struct {
	Re     *regexp.Regexp
	Action string
}

func LoadPatterns() []CompiledPattern {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../..")
	viper.AddConfigPath("..")

	if err := viper.ReadInConfig(); err != nil {
		// No config found
	}

	var configs []PatternConfig
	viper.UnmarshalKey("patterns", &configs)

	if len(configs) == 0 {
		fmt.Println("Warning: No patterns found in config.yaml")
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
