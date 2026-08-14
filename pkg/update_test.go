package pkg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Regression test: pkg.Update() calls every operation's Update() in a
// location and writes back whatever it returns. Before the no-op
// operations were fixed to return their current state unchanged, they
// unconditionally returned "", so introducing any operation that actually
// needs Update() called (e.g. Cycle) would silently blank out state other
// operations set via `set-state` the moment `commandline_thing update` ran
// for that location.
func TestUpdatePreservesOtherOperationsStateAlongsideCycle(t *testing.T) {
	store := NewMemoryStateStore()
	require.NoError(t, store.Set("prompt", "12345", "exit_code", "1"))
	require.NoError(t, store.Set("prompt", "12345", "vim", "1"))

	config := Location{
		Operations: []OperationWrapper{
			{Operation: &ExitCode{}},
			{Operation: &VimMode{}},
			{Operation: mustConfiguredCycle(t, "nyan", "nyan1", "nyan2")},
		},
	}

	require.NoError(t, Update(store, config, "prompt", "12345", "/tmp"))

	exitCode, err := store.Get("prompt", "12345", "exit_code")
	require.NoError(t, err)
	require.Equal(t, "1", exitCode, "exit_code state must survive an Update() call for an unrelated operation")

	vim, err := store.Get("prompt", "12345", "vim")
	require.NoError(t, err)
	require.Equal(t, "1", vim, "vim state must survive an Update() call for an unrelated operation")

	nyan, err := store.Get("prompt", "12345", "nyan")
	require.NoError(t, err)
	require.Equal(t, "1", nyan, "cycle's own state must advance")
}

func mustConfiguredCycle(t *testing.T, name string, names ...string) *Cycle {
	t.Helper()
	c := &Cycle{}
	rawNames := make([]interface{}, len(names))
	for i, n := range names {
		rawNames[i] = n
	}
	require.NoError(t, c.Configure(map[string]interface{}{
		"name":  name,
		"names": rawNames,
	}))
	return c
}
