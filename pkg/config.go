package pkg

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type PaneConfig struct {
	Separator               string                   `json:"separator" yaml:"separator"`
	SeparatorColor          string                   `json:"separator_color" yaml:"separator_color"`
	BackgroundColor         string                   `json:"background_color" yaml:"background_color"`
	OperatorBackgroundColor string                   `json:"operator_background_color" yaml:"operator_background_color"`
	OperationsLeftRaw       []map[string]interface{} `json:"operations_left" yaml:"operations_left"`
	OperationsRightRaw      []map[string]interface{} `json:"operations_right" yaml:"operations_right"`
	OperationsLeft          []Operation
	OperationsRight         []Operation
}

type Config struct {
	PaneConfig PaneConfig `json:"pane_config"`
	// PromptOperations []Operation
}

func (config *Config) GeneratePaneLeft() (string, error) {
	var pane strings.Builder
	for _, op := range config.PaneConfig.OperationsLeft {
		pane.WriteString(" | ")
		opOutput, err := op.Generate(nil)
		if err != nil {
			return "", err
		}
		pane.WriteString(opOutput)
	}

	return pane.String(), nil
}

func (config *Config) GeneratePaneRight() (string, error) {
	var pane strings.Builder
	for _, op := range config.PaneConfig.OperationsRight {
		pane.WriteString(" | ")
		opOutput, err := op.Generate(nil)
		if err != nil {
			return "", err
		}
		pane.WriteString(opOutput)
	}

	return pane.String(), nil
}

func LoadConfig(availableOperations map[string]OperationLoader) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/commandline_thing")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("No config file found, create one at $HOME/.config/commandline_thing")
	}

	var config *Config
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	for _, opRaw := range config.PaneConfig.OperationsLeftRaw {
		var typRaw interface{}

		if typRaw, ok := opRaw["type"]; !ok {
			return nil, fmt.Errorf("Error with config: %v", typRaw)
		}

		typ, ok := typRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Type is unknown type: %v", typRaw)
		}

		var opLoader OperationLoader
		if opLoader, ok = availableOperations[typ]; !ok {
			return nil, fmt.Errorf("Unknown Operation: %v", typ)
		}

		op, err := opLoader(opRaw)
		if err != nil {
			return nil, err
		}

		config.PaneConfig.OperationsLeft = append(config.PaneConfig.OperationsLeft, op)
	}

	for _, opRaw := range config.PaneConfig.OperationsRightRaw {
		var typRaw interface{}

		if typRaw, ok := opRaw["type"]; !ok {
			return nil, fmt.Errorf("Error with config: %v", typRaw)
		}

		typ, ok := typRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Type is unknown type: %v", typRaw)
		}

		var opLoader OperationLoader
		if opLoader, ok = availableOperations[typ]; !ok {
			return nil, fmt.Errorf("Unknown Operation: %v", typ)
		}
		op, err := opLoader(opRaw)
		if err != nil {
			return nil, err
		}
		config.PaneConfig.OperationsRight = append(config.PaneConfig.OperationsLeft, op)
	}

	return config, nil
}
