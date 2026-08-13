package pkg

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MemeDirEnvVar is checked before falling back to MemeDirDefault, matching
// the `meme` zsh function in dotfiles (meme.sh) so both implementations
// resolve to the same directory unless overridden.
const MemeDirEnvVar = "MEME_DIR"

// MemeDirDefault mirrors meme.sh's default.
const MemeDirDefault = "dotfiles/nix/memes"

// MemeDir resolves the directory memes are read from: $MEME_DIR if set,
// otherwise ~/dotfiles/nix/memes.
func MemeDir() string {
	if dir := os.Getenv(MemeDirEnvVar); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return MemeDirDefault
	}
	return filepath.Join(home, MemeDirDefault)
}

// ListMemes returns the available meme names in dir (filenames with their
// extension stripped), sorted and deduplicated.
func ListMemes(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading meme dir %s: %w", dir, err)
	}
	seen := make(map[string]bool)
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// MemeCodepointBase is U+100000, the start of the Supplementary Private Use
// Area (Plane 16). MemeTerminal (the iTerm2 fork at github:pj/iTerm2, branch
// memeterminal) patches this 1024-code-point range to render double-width
// and pulls its glyphs from MemeFont.ttf's SBIX color bitmaps.
const MemeCodepointBase = 0x100000

// MemeCodepoint returns the PUA code point assigned to name: MemeCodepointBase
// plus name's index in ListMemes' sorted output. This is the same order
// emojifont's `just build-dotfiles-font` recipe uses to assign code points
// when injecting glyphs into MemeFont.ttf, so the two stay in sync as long
// as the font is rebuilt whenever the meme directory's contents change —
// there's no separate manifest file recording the mapping.
func MemeCodepoint(dir, name string) (rune, error) {
	names, err := ListMemes(dir)
	if err != nil {
		return 0, err
	}
	for i, n := range names {
		if n == name {
			return rune(MemeCodepointBase + i), nil
		}
	}
	return 0, fmt.Errorf("no meme named %q in %s", name, dir)
}

// ResolveMeme finds the file in dir for the given name. It matches the exact
// filename first (so "pepe.jpg" works even if invoked with the extension),
// then falls back to a case-insensitive match on filename-without-extension.
// Multiple matches (e.g. both pepe.jpg and pepe.png present) is an error
// rather than silently picking one, matching meme.sh's behavior.
func ResolveMeme(dir, name string) (string, error) {
	exact := filepath.Join(dir, name)
	if info, err := os.Stat(exact); err == nil && !info.IsDir() {
		return exact, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("reading meme dir %s: %w", dir, err)
	}

	lowerName := strings.ToLower(name)
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if strings.ToLower(stem) == lowerName {
			matches = append(matches, e.Name())
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no meme named %q in %s", name, dir)
	case 1:
		return filepath.Join(dir, matches[0]), nil
	default:
		return "", fmt.Errorf("%q is ambiguous, matches: %s", name, strings.Join(matches, ", "))
	}
}

// BuildEscapeSequence reads the image at path and returns the iTerm2 OSC 1337
// File= inline-image escape sequence for it, sized to width x height cells.
// This is the same protocol `imgcat`/it2cli use, and the same one MemeTerminal
// (and any other iTerm2-family terminal) already supports natively — no
// custom font or font injection needed.
func BuildEscapeSequence(path string, width, height int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf(
		"\033]1337;File=inline=1;width=%d;height=%d;preserveAspectRatio=0:%s\a",
		width, height, encoded,
	), nil
}

// WrapForTmux wraps seq in a DCS passthrough sequence so it survives tmux
// instead of being swallowed. tmux requires `set -g allow-passthrough on`
// for this to actually reach the outer terminal (see dotfiles/nix/tmux.conf).
// Any ESC byte already in seq must be doubled per the DCS passthrough spec.
func WrapForTmux(seq string) string {
	doubled := strings.ReplaceAll(seq, "\033", "\033\033")
	return "\033Ptmux;" + doubled + "\033\\"
}

// BuildMemeSequence resolves name to a file in dir and returns the
// inline-image escape sequence at width x height cells, wrapped for tmux
// passthrough if TMUX is set in the environment.
func BuildMemeSequence(dir, name string, width, height int) (string, error) {
	path, err := ResolveMeme(dir, name)
	if err != nil {
		return "", err
	}
	seq, err := BuildEscapeSequence(path, width, height)
	if err != nil {
		return "", err
	}
	if os.Getenv("TMUX") != "" {
		seq = WrapForTmux(seq)
	}
	return seq, nil
}

// EmitMeme resolves name to a file in dir, builds the inline-image escape
// sequence at width x height cells, and writes it to w — wrapped for tmux
// passthrough if TMUX is set in the environment.
func EmitMeme(w io.Writer, dir, name string, width, height int) error {
	seq, err := BuildMemeSequence(dir, name, width, height)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, seq)
	return err
}
