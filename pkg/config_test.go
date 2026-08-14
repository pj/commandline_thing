package pkg

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperationWrapperDecodeHookConfiguresOperation(t *testing.T) {
	availableOperations = LoadAvailableOperations()

	hook := OperationWrapperDecodeHook()
	raw := map[string]interface{}{
		"type":  "cycle",
		"name":  "nyan",
		"names": []interface{}{"nyan1", "nyan2", "nyan3", "nyan4"},
	}

	result, err := hook(reflect.TypeOf(raw), reflect.TypeOf(OperationWrapper{}), raw)
	require.NoError(t, err)

	wrapper, ok := result.(*OperationWrapper)
	require.True(t, ok)
	require.Equal(t, OperationName("nyan"), wrapper.Operation.Name())

	cycle, ok := wrapper.Operation.(*Cycle)
	require.True(t, ok)
	require.Equal(t, []string{"nyan1", "nyan2", "nyan3", "nyan4"}, cycle.names)
}

func TestOperationWrapperDecodeHookPropagatesConfigureError(t *testing.T) {
	availableOperations = LoadAvailableOperations()

	hook := OperationWrapperDecodeHook()
	raw := map[string]interface{}{"type": "cycle"} // missing required "names"

	_, err := hook(reflect.TypeOf(raw), reflect.TypeOf(OperationWrapper{}), raw)
	require.Error(t, err)
	require.Contains(t, err.Error(), "names is required")
}

func TestOperationWrapperDecodeHookSkipsConfigureForNonConfigurableOperations(t *testing.T) {
	availableOperations = LoadAvailableOperations()

	hook := OperationWrapperDecodeHook()
	raw := map[string]interface{}{"type": "git"}

	result, err := hook(reflect.TypeOf(raw), reflect.TypeOf(OperationWrapper{}), raw)
	require.NoError(t, err)

	wrapper, ok := result.(*OperationWrapper)
	require.True(t, ok)
	require.Equal(t, OperationName("git"), wrapper.Operation.Name())
}
