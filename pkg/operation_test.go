package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestMemeOperationGenerateReturnsCodepointPerName(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "doge.png"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pepe.jpg"), []byte("b"), 0644))
	t.Setenv(MemeDirEnvVar, dir)

	op := &Meme{}
	require.Equal(t, OperationName("meme"), op.Name())

	result, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	memes, ok := result.(map[string]string)
	require.True(t, ok)
	// ListMemes sorts "doge" before "pepe", so doge gets the base codepoint.
	require.Equal(t, string(rune(MemeCodepointBase)), memes["doge"])
	require.Equal(t, string(rune(MemeCodepointBase+1)), memes["pepe"])
}

func TestMemeOperationUsableAsDottedTemplateField(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pepe.jpg"), []byte("hello"), 0644))
	t.Setenv(MemeDirEnvVar, dir)

	op := &Meme{}
	result, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	tmpl, err := template.New("t").Parse("{{ .meme.pepe }}")
	require.NoError(t, err)

	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, map[string]interface{}{"meme": result}))
	require.Equal(t, string(rune(MemeCodepointBase)), buf.String())
}

func TestMemeOperationRegisteredInAvailableOperations(t *testing.T) {
	ops := LoadAvailableOperations()
	newOp, ok := ops["meme"]
	require.True(t, ok)
	require.IsType(t, &Meme{}, newOp())
}

func TestCycleRegisteredInAvailableOperations(t *testing.T) {
	ops := LoadAvailableOperations()
	newOp, ok := ops["cycle"]
	require.True(t, ok)
	require.IsType(t, &Cycle{}, newOp())
	// An unconfigured Cycle must report the type name "cycle" — that's the
	// registry key above, and also what a bare `type: cycle` YAML entry
	// matches against before Configure() ever runs.
	require.Equal(t, OperationName("cycle"), (&Cycle{}).Name())
}

func TestCycleConfigureSetsNameAndNames(t *testing.T) {
	c := &Cycle{}
	err := c.Configure(map[string]interface{}{
		"type":  "cycle",
		"name":  "nyan",
		"names": []interface{}{"nyan1", "nyan2", "nyan3", "nyan4"},
	})
	require.NoError(t, err)
	require.Equal(t, OperationName("nyan"), c.Name())
}

func TestCycleConfigureWithoutNameKeepsTypeAsName(t *testing.T) {
	c := &Cycle{}
	err := c.Configure(map[string]interface{}{
		"type":  "cycle",
		"names": []interface{}{"a", "b"},
	})
	require.NoError(t, err)
	require.Equal(t, OperationName("cycle"), c.Name())
}

func TestCycleConfigureRequiresNames(t *testing.T) {
	c := &Cycle{}
	err := c.Configure(map[string]interface{}{"type": "cycle", "name": "nyan"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "names is required")
}

func TestCycleConfigureRejectsEmptyNames(t *testing.T) {
	c := &Cycle{}
	err := c.Configure(map[string]interface{}{"names": []interface{}{}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "must not be empty")
}

func TestCycleConfigureRejectsNonStringNames(t *testing.T) {
	c := &Cycle{}
	err := c.Configure(map[string]interface{}{"names": []interface{}{"a", 2}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "list of strings")
}

func TestCycleUpdateAdvancesAndWraps(t *testing.T) {
	c := &Cycle{}
	require.NoError(t, c.Configure(map[string]interface{}{
		"names": []interface{}{"a", "b", "c"},
	}))

	next, err := c.Update("", "")
	require.NoError(t, err)
	require.Equal(t, "1", next)

	next, err = c.Update("", "1")
	require.NoError(t, err)
	require.Equal(t, "2", next)

	// wraps back to 0 after the last index
	next, err = c.Update("", "2")
	require.NoError(t, err)
	require.Equal(t, "0", next)
}

func TestCycleUpdateTreatsInvalidStateAsZero(t *testing.T) {
	c := &Cycle{}
	require.NoError(t, c.Configure(map[string]interface{}{
		"names": []interface{}{"a", "b"},
	}))

	next, err := c.Update("", "not-a-number")
	require.NoError(t, err)
	require.Equal(t, "1", next)
}

func TestCycleGenerateReturnsNameAtState(t *testing.T) {
	c := &Cycle{}
	require.NoError(t, c.Configure(map[string]interface{}{
		"names": []interface{}{"nyan1", "nyan2", "nyan3", "nyan4"},
	}))

	result, err := c.Generate("prompt", "12345", "", "2")
	require.NoError(t, err)
	require.Equal(t, "nyan3", result)
}

func TestCycleGenerateOutOfRangeStateFallsBackToZero(t *testing.T) {
	c := &Cycle{}
	require.NoError(t, c.Configure(map[string]interface{}{
		"names": []interface{}{"a", "b"},
	}))

	result, err := c.Generate("prompt", "12345", "", "99")
	require.NoError(t, err)
	require.Equal(t, "a", result)
}

func TestCycleUsableWithMemeMapInTemplateViaIndex(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nyan1.png"), []byte("a"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nyan2.png"), []byte("b"), 0644))
	t.Setenv(MemeDirEnvVar, dir)

	memeOp := &Meme{}
	memes, err := memeOp.Generate("prompt", "12345", dir, "")
	require.NoError(t, err)

	cycleOp := &Cycle{}
	require.NoError(t, cycleOp.Configure(map[string]interface{}{
		"name":  "nyan",
		"names": []interface{}{"nyan1", "nyan2"},
	}))
	current, err := cycleOp.Generate("prompt", "12345", dir, "1")
	require.NoError(t, err)

	tmpl, err := template.New("t").Parse(`{{ index .meme .nyan }}`)
	require.NoError(t, err)

	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, map[string]interface{}{"meme": memes, "nyan": current}))
	require.Equal(t, string(rune(MemeCodepointBase+1)), buf.String())
}
