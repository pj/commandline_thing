package pkg

import (
	"bytes"
	"fmt"
	"text/template"
)

// GenerateContent takes a LocationConfig and generates content based on the operations and template
func GenerateContent(state StateStore, config Location, locationKey LocationKey, instanceKey InstanceKey, locationPath string) (string, error) {
	// Create a map to store data from operations
	data := make(map[string]interface{})

	// Load data from each operation
	for _, opWrapper := range config.Operations {
		op := opWrapper.Operation
		operationName := op.Name()
		operationState, err := state.Get(locationKey, instanceKey, operationName)
		if err != nil {
			return "", fmt.Errorf("error getting state for operation %s: %w", op.Name(), err)
		}
		// var operationStateInterface interface{}
		// if err := json.Unmarshal([]byte(operationState), &operationStateInterface); err != nil {
		// 	return "", fmt.Errorf("error unmarshaling state for operation %s: %w", op.Name(), err)
		// }
		result, err := op.Generate(locationKey, instanceKey, locationPath, operationState)
		if err != nil {
			return "", fmt.Errorf("error generating data for operation %s: %w", op.Name(), err)
		}
		data[string(op.Name())] = result
	}

	// Parse the template
	tmpl, err := template.New("content").Parse(config.Template)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	// Execute the template with the data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
}
