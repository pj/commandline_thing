package pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestResolveMemeExactFilename(t *testing.T) {
	dir := t.TempDir()
	path := writeTestImage(t, dir, "pepe.jpg", []byte("a"))

	got, err := ResolveMeme(dir, "pepe.jpg")
	require.NoError(t, err)
	require.Equal(t, path, got)
}

func TestResolveMemeByNameWithoutExtension(t *testing.T) {
	dir := t.TempDir()
	path := writeTestImage(t, dir, "pepe.jpg", []byte("a"))

	got, err := ResolveMeme(dir, "pepe")
	require.NoError(t, err)
	require.Equal(t, path, got)
}

func TestResolveMemeCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	path := writeTestImage(t, dir, "GigaChad.png", []byte("a"))

	got, err := ResolveMeme(dir, "gigachad")
	require.NoError(t, err)
	require.Equal(t, path, got)
}

func TestResolveMemeNotFound(t *testing.T) {
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))

	_, err := ResolveMeme(dir, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no meme named")
}

func TestResolveMemeAmbiguous(t *testing.T) {
	// Two different files, same stem, different extension — this must not
	// silently pick one; that's how a font-injection sibling project
	// (emojifont) originally decided this exact ambiguity should fail loud.
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("a"))
	writeTestImage(t, dir, "pepe.png", []byte("b"))

	_, err := ResolveMeme(dir, "pepe")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous")
}

func TestBuildEscapeSequenceStructure(t *testing.T) {
	dir := t.TempDir()
	path := writeTestImage(t, dir, "pepe.jpg", []byte("hello"))

	seq, err := BuildEscapeSequence(path, 2, 1)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(seq, "\033]1337;File=inline=1;width=2;height=1;preserveAspectRatio=0:"))
	require.True(t, strings.HasSuffix(seq, "\a"))
	// base64("hello") = aGVsbG8=
	require.Contains(t, seq, "aGVsbG8=")
}

func TestBuildEscapeSequenceMissingFile(t *testing.T) {
	_, err := BuildEscapeSequence("/nonexistent/path.jpg", 2, 1)
	require.Error(t, err)
}

func TestWrapForTmuxDoublesEscapeBytes(t *testing.T) {
	// Verified byte-for-byte against a live iTerm2/tmux test during
	// development; this locks in that exact structure:
	//   ESC P t m u x ; <payload with every ESC doubled> ESC \
	seq := "\033]1337;File=x\a"
	wrapped := WrapForTmux(seq)

	require.True(t, strings.HasPrefix(wrapped, "\033Ptmux;"))
	require.True(t, strings.HasSuffix(wrapped, "\033\\"))

	// Between the "tmux;" header and the trailing "ESC \", every ESC from
	// the original sequence must appear doubled.
	inner := strings.TrimSuffix(strings.TrimPrefix(wrapped, "\033Ptmux;"), "\033\\")
	require.Equal(t, strings.ReplaceAll(seq, "\033", "\033\033"), inner)
}

func TestWrapForTmuxNoEscBytesIsUnchangedInside(t *testing.T) {
	wrapped := WrapForTmux("plain text, no escapes")
	require.Equal(t, "\033Ptmux;plain text, no escapes\033\\", wrapped)
}

func TestEmitMemeWritesSequence(t *testing.T) {
	t.Setenv("TMUX", "")
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("hello"))

	var buf bytes.Buffer
	err := EmitMeme(&buf, dir, "pepe", 2, 1)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "aGVsbG8=")
	require.False(t, strings.HasPrefix(buf.String(), "\033Ptmux;"))
}

func TestEmitMemeWrapsForTmuxWhenSet(t *testing.T) {
	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	dir := t.TempDir()
	writeTestImage(t, dir, "pepe.jpg", []byte("hello"))

	var buf bytes.Buffer
	err := EmitMeme(&buf, dir, "pepe", 2, 1)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(buf.String(), "\033Ptmux;"))
	require.True(t, strings.HasSuffix(buf.String(), "\033\\"))
}

func TestEmitMemeUnknownName(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := EmitMeme(&buf, dir, "nope", 2, 1)
	require.Error(t, err)
	require.Empty(t, buf.String())
}
