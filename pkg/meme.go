package pkg

import (
	"fmt"
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

// MemeCodepoint resolves name to a meme in dir — the exact ListMemes() name
// first, falling back to a case-insensitive match (erroring if that's
// ambiguous, e.g. both "Pepe" and "pepe" present as distinct entries) — and
// returns its PUA code point: MemeCodepointBase plus its index in ListMemes'
// sorted output. This is the same order emojifont's build_meme_mappings_from_dir
// (used by the Nix derivation that builds MemeFont.ttf, see dotfiles/nix/flake.nix)
// assigns code points in, so the two stay in sync as long as the font is
// rebuilt whenever the meme directory's contents change — there's no
// separate manifest file recording the mapping.
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

	lowerName := strings.ToLower(name)
	matchIdx := -1
	var ambiguous []string
	for i, n := range names {
		if strings.ToLower(n) == lowerName {
			if matchIdx != -1 {
				ambiguous = append(ambiguous, names[matchIdx], n)
			}
			matchIdx = i
		}
	}
	if len(ambiguous) > 0 {
		return 0, fmt.Errorf("%q is ambiguous, matches: %s", name, strings.Join(ambiguous, ", "))
	}
	if matchIdx == -1 {
		return 0, fmt.Errorf("no meme named %q in %s", name, dir)
	}
	return rune(MemeCodepointBase + matchIdx), nil
}
