package config

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/spf13/viper"
)

type RuleConfig struct {
	ID          int    `mapstructure:"id" json:"id"`
	Regex       string `mapstructure:"regex" json:"regex"`
	Action      string `mapstructure:"action" json:"action"`
	Description string `mapstructure:"description" json:"description"`
}

type CompiledRule struct {
	Re     *regexp.Regexp
	Action string
}

func LoadRules() []CompiledRule {
	configs, err := LoadConfigs(nil)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	var compiledRules []CompiledRule
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			fmt.Printf("Error compiling regex '%s': %s\n", cfg.Regex, err)
			continue
		}
		compiledRules = append(compiledRules, CompiledRule{Re: re, Action: cfg.Action})
	}
	return compiledRules
}

// LoadRulesWithConfig allows loading from specific bytes (e.g. embedded)
func LoadRulesWithConfig(configBytes []byte) []CompiledRule {
	configs, err := LoadConfigs(configBytes)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	var compiledRules []CompiledRule
	for _, cfg := range configs {
		re, err := regexp.Compile(cfg.Regex)
		if err != nil {
			fmt.Printf("Error compiling regex '%s': %s\n", cfg.Regex, err)
			continue
		}
		compiledRules = append(compiledRules, CompiledRule{Re: re, Action: cfg.Action})
	}
	return compiledRules
}

func LoadConfigs(configBytes []byte) ([]RuleConfig, error) {
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

	var configs []RuleConfig
	err := viper.UnmarshalKey("rules", &configs)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling rules: %w", err)
	}
	return configs, nil
}

type TestCase struct {
	ID           int                    `mapstructure:"id" json:"id"`
	Description  string                 `mapstructure:"description" json:"description"`
	Input        map[string]interface{} `mapstructure:"input" json:"input"`
	Text         string                 `mapstructure:"text" json:"text"`
	Expected     string                 `mapstructure:"expected" json:"expected"`
	CsvExpected  string                 `mapstructure:"csv_expected" json:"csv_expected"`
	Passed       bool                   `mapstructure:"passed" json:"passed"`
	SkipFromTest bool                   `mapstructure:"skip_from_test" json:"skip_from_test"`
}

func LoadTests(configBytes []byte) []TestCase {
	viper.Reset()
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			fmt.Printf("Error reading test config bytes: %v\n", err)
			return nil
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		viper.AddConfigPath("..")
		// Also add module root if running from pkg/config
		viper.AddConfigPath("../../..")

		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Error reading config for tests: %v\n", err)
		}
	}

	var tests []TestCase
	err := viper.UnmarshalKey("test", &tests)
	if err != nil {
		fmt.Printf("Error unmarshalling tests: %v\n", err)
	}
	return tests
}

func LoadBuildInfo(configBytes []byte) (string, string) {
	viper.Reset()
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return "unknown", "unknown"
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		// Also add module root if running from pkg/config
		viper.AddConfigPath("../../..")

		if err := viper.ReadInConfig(); err != nil {
			return "unknown", "unknown"
		}
	}

	return viper.GetString("application.version"), viper.GetString("application.last_build")
}

type AppConfig struct {
	Name               string `mapstructure:"name"`
	Version            string `mapstructure:"version"`
	LastBuild          string `mapstructure:"last_build"`
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	ReferenceDocPath   string `mapstructure:"reference_doc_path"`
	ParserRulesDocPath string `mapstructure:"parser_rules_doc_path"`
	SqlExecuteMode     string `mapstructure:"sql_execute_mode"`
	SqlLimits          []int  `mapstructure:"sql_limits"`
}

func LoadAppConfig(configBytes []byte) AppConfig {
	viper.Reset()
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return AppConfig{}
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		viper.AddConfigPath("../../..")
		if err := viper.ReadInConfig(); err != nil {
			return AppConfig{}
		}
	}

	var app AppConfig
	err := viper.UnmarshalKey("application", &app)
	if err != nil {
		fmt.Printf("Error unmarshalling application config: %v\n", err)
	}
	if app.SqlExecuteMode == "" {
		app.SqlExecuteMode = "EXECUTE"
	}
	if len(app.SqlLimits) == 0 {
		app.SqlLimits = []int{10, 25, 50, 100}
	}
	return app
}

type DatabaseConfig struct {
	Name     string `mapstructure:"name" json:"name"`
	Host     string `mapstructure:"host" json:"host"`
	Port     string `mapstructure:"port" json:"port"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	Database string `mapstructure:"database" json:"database"`
	Schema   string `mapstructure:"schema" json:"schema"`
}

func LoadDatabaseConfigs(configBytes []byte) []DatabaseConfig {
	viper.Reset()
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return nil
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		viper.AddConfigPath("../../..")
		if err := viper.ReadInConfig(); err != nil {
			return nil
		}
	}

	var dbs []DatabaseConfig
	err := viper.UnmarshalKey("databases", &dbs)
	if err != nil {
		fmt.Printf("Error unmarshalling databases config: %v\n", err)
	}
	return dbs
}

type CSVConfig struct {
	Format    string `mapstructure:"format"`
	Header    bool   `mapstructure:"header"`
	Delimiter string `mapstructure:"delimiter"`
	Null      string `mapstructure:"null"`
	Encoding  string `mapstructure:"encoding"`
	Quote     string `mapstructure:"quote"`
}

func LoadCSVConfig(configBytes []byte) CSVConfig {
	viper.Reset()
	viper.SetConfigType("yaml")

	if configBytes != nil {
		if err := viper.ReadConfig(bytes.NewBuffer(configBytes)); err != nil {
			return CSVConfig{}
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("../..")
		viper.AddConfigPath("../../..")
		if err := viper.ReadInConfig(); err != nil {
			return CSVConfig{
				Format:    "CSV",
				Header:    true,
				Delimiter: ";",
				Null:      "N/A",
				Encoding:  "UTF8",
				Quote:     "\"",
			} // Fallback defaults
		}
	}

	var csv CSVConfig
	err := viper.UnmarshalKey("copy.csv", &csv)
	if err != nil {
		fmt.Printf("Error unmarshalling copy.csv config: %v\n", err)
	}
	return csv
}
