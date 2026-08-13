package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

func TestMemeOperationGenerateReturnsSequencePerName(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pepe.jpg"), []byte("hello"), 0644))
	t.Setenv(MemeDirEnvVar, dir)

	op := &Meme{}
	require.Equal(t, OperationName("meme"), op.Name())

	result, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	memes, ok := result.(map[string]string)
	require.True(t, ok)
	require.Contains(t, memes, "pepe")
	// base64("hello") = aGVsbG8=
	require.Contains(t, memes["pepe"], "aGVsbG8=")
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
	require.Contains(t, buf.String(), "aGVsbG8=")
}

func TestMemeOperationRegisteredInAvailableOperations(t *testing.T) {
	ops := LoadAvailableOperations()
	newOp, ok := ops["meme"]
	require.True(t, ok)
	require.IsType(t, &Meme{}, newOp())
}
