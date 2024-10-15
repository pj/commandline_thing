package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

type OperationWrapper struct {
	Operation Operation
}

var availableOperations Operations

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (ow *OperationWrapper) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawOp map[string]interface{}
	if err := unmarshal(&rawOp); err != nil {
		return err
	}

	typRaw, ok := rawOp["type"]
	if !ok {
		return fmt.Errorf("operation type not specified")
	}

	typ, ok := typRaw.(string)
	if !ok {
		return fmt.Errorf("operation type is not a string")
	}

	opName := OperationName(typ)
	newOperation, ok := availableOperations[opName]
	if !ok {
		return fmt.Errorf("unknown operation type: %s", typ)
	}

	op := newOperation()

	ow.Operation = op
	return nil
}

type Location struct {
	Operations []OperationWrapper `mapstructure:"operations"`
	Template   string             `mapstructure:"template"`
}

type AllConfigs struct {
	Configs map[LocationKey]Location `mapstructure:"configs"`
}

func LoadConfig(loadedOperations Operations) (*AllConfigs, error) {
	availableOperations = loadedOperations
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/commandline_thing")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("no config file found, create one at $HOME/.config/commandline_thing")
	}

	var config *AllConfigs
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
