package pkg

import "fmt"

func Update(stateStore StateStore, config Location, locationKey LocationKey, instanceKey InstanceKey, locationPath string) error {
	for _, opWrapper := range config.Operations {
		op := opWrapper.Operation
		operationName := op.Name()
		operationState, err := stateStore.Get(locationKey, instanceKey, operationName)
		if err != nil {
			return fmt.Errorf("error getting state for operation %s: %w", op.Name(), err)
		}
		nextState, err := op.Update(locationPath, operationState)
		if err != nil {
			return fmt.Errorf("error generating data for operation %s: %w", op.Name(), err)
		}

		err = stateStore.Set(locationKey, instanceKey, operationName, nextState)
		if err != nil {
			return fmt.Errorf("error setting state for operation %s: %w", op.Name(), err)
		}
	}

	return nil
}
