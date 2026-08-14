package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTestImage(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, content, 0644))
	return path
}

func TestMemeDirEnvOverride(t *testing.T) {
	t.Setenv(MemeDirEnvVar, "/custom/meme/dir")
	require.Equal(t, "/custom/meme/dir", MemeDir())
}

func TestMemeDirDefault(t *testing.T) {
	t.Setenv(MemeDirEnvVar, "")
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, MemeDirDefault), MemeDir())
}

func TestListMemes(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))
	writeTestImage(t, dir, "doge.png", []byte("b"))

	names, err := ListMemes(dir)
	require.NoError(t, err)
	require.Equal(t, []string{"doge", "pepe"}, names)
}

func TestListMemesDedupesAcrossExtensions(t *testing.T) {
	// Same stem, different extension — ListMemes just enumerates names for
	// display; ambiguity is ResolveMeme's problem, not this function's.
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))
	writeTestImage(t, dir, "pepe.png", []byte("b"))

	names, err := ListMemes(dir)
	require.NoError(t, err)
	require.Equal(t, []string{"pepe"}, names)
}

func TestListMemesIgnoresDirectories(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0755))

	names, err := ListMemes(dir)
	require.NoError(t, err)
	require.Equal(t, []string{"pepe"}, names)
}

func TestMemeCodepointAssignsByIndexInSortedOrder(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "doge.png", []byte("a"))
	writeTestImage(t, dir, "pepe.jpg", []byte("b"))

	got, err := MemeCodepoint(dir, "doge")
	require.NoError(t, err)
	require.Equal(t, rune(MemeCodepointBase), got)

	got, err = MemeCodepoint(dir, "pepe")
	require.NoError(t, err)
	require.Equal(t, rune(MemeCodepointBase+1), got)
}

func TestMemeCodepointNotFound(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))

	_, err := MemeCodepoint(dir, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no meme named")
}

func TestMemeCodepointCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "GigaChad.png", []byte("a"))

	got, err := MemeCodepoint(dir, "gigachad")
	require.NoError(t, err)
	require.Equal(t, rune(MemeCodepointBase), got)
}

func TestMemeCodepointAmbiguous(t *testing.T) {
	// Two distinct entries differing only by case — ListMemes dedupes by
	// exact stem, so both "Pepe" and "pepe" survive as separate names; a
	// case-insensitive lookup of either must not silently pick one.
	dir := t.TempDir()
	writeTestImage(t, dir, "Pepe.jpg", []byte("a"))
	writeTestImage(t, dir, "pepe.png", []byte("b"))

	_, err := MemeCodepoint(dir, "PEPE")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous")
}
