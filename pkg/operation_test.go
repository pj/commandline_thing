package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

// requireBinary skips the test when name isn't on PATH, rather than failing
// a machine that simply doesn't have git/jj installed.
func requireBinary(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not found on PATH", name)
	}
}

// runCmd runs name with args in dir, failing the test on any error. Used to
// set up real git/jj repos for the Git operation's tests below — this
// codebase's existing tests (see Meme below) already favor exercising real
// on-disk/subprocess behavior over mocking, and there's no exec.Command
// abstraction here to mock even if we wanted to.
func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "%s %s: %s", name, strings.Join(args, " "), output)
}

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

func TestGitOperationRegisteredInAvailableOperations(t *testing.T) {
	ops := LoadAvailableOperations()
	newOp, ok := ops["git"]
	require.True(t, ok)
	require.IsType(t, &Git{}, newOp())
}

func TestGitOperationReturnsGitBranchAndCleanStatus(t *testing.T) {
	requireBinary(t, "git")
	dir := t.TempDir()
	runCmd(t, dir, "git", "init", "-b", "main", "-q")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hello"), 0644))
	runCmd(t, dir, "git", "add", "f.txt")
	runCmd(t, dir, "git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-qm", "init")

	op := &Git{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	require.Equal(t, GitResult{Branch: "main", Status: ""}, generated)
}

func TestGitOperationReportsDirtyStatus(t *testing.T) {
	requireBinary(t, "git")
	dir := t.TempDir()
	runCmd(t, dir, "git", "init", "-b", "main", "-q")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hello"), 0644))
	runCmd(t, dir, "git", "add", "f.txt")
	runCmd(t, dir, "git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-qm", "init")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f.txt"), []byte("changed"), 0644))

	op := &Git{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	result, ok := generated.(GitResult)
	require.True(t, ok)
	require.NotEmpty(t, result.Status, "a modified tracked file must show as dirty via git status -s")
}

func TestGitOperationReturnsNilOutsideAnyRepo(t *testing.T) {
	dir := t.TempDir()

	op := &Git{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")

	require.NoError(t, err)
	require.Nil(t, generated, "no repo present — nothing to report")
}

func TestJjOperationRegisteredInAvailableOperations(t *testing.T) {
	ops := LoadAvailableOperations()
	newOp, ok := ops["jj"]
	require.True(t, ok)
	require.IsType(t, &Jj{}, newOp())
}

func TestJjOperationReturnsNilOutsideAJJRepo(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")

	require.NoError(t, err)
	require.Nil(t, generated, "an empty directory with no .jj must not be treated as a jj repo")
}

func TestJjOperationReturnsNilWhenJJBinaryMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", "")

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")

	require.NoError(t, err)
	require.Nil(t, generated)
}

func TestJjOperationUsesChangeIDWhenAtHasNoBookmark(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()
	runCmd(t, dir, "jj", "git", "init")

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	result, ok := generated.(JjResult)
	require.True(t, ok)
	require.Empty(t, result.Status, "a freshly initialized jj repo's @ has no diff from its parent yet")
	require.GreaterOrEqual(t, len(result.Branch), 3, "with no bookmark at @, Branch should fall back to the minimal unique change-id prefix (shortest(3))")
}

func TestJjOperationShowsClosestBookmarkAndChangeIDWhenOnTopOfBookmark(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()
	runCmd(t, dir, "jj", "git", "init")
	runCmd(t, dir, "jj", "bookmark", "create", "main", "-r", "@-")
	runCmd(t, dir, "jj", "new")

	cmd := exec.Command("jj", "log", "-r", "@", "--no-graph", "-T", "change_id.shortest(3)")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	result, ok := generated.(JjResult)
	require.True(t, ok)
	require.Equal(t, "main > #"+strings.TrimSpace(string(out)), result.Branch)
	require.Empty(t, result.Bookmark, "@ has no bookmark of its own here")
	require.Equal(t, "main", result.LastBookmark)

	changeCmd := exec.Command("jj", "log", "-r", "@", "--no-graph", "-T", `change_id ++ "\x1f" ++ commit_id ++ "\x1f" ++ commit_id.short(8)`)
	changeCmd.Dir = dir
	idsOut, err := changeCmd.Output()
	require.NoError(t, err)
	ids := strings.Split(strings.TrimSpace(string(idsOut)), "\x1f")
	require.Len(t, ids, 3)
	require.Equal(t, ids[0], result.ChangeIDFull)
	require.Equal(t, strings.TrimSpace(string(out)), result.ChangeIDShort)
	require.Equal(t, ids[1], result.GitIDFull)
	require.Equal(t, ids[2], result.GitIDShort)
}

func TestJjOperationReportsBookmarkAsBranch(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()
	runCmd(t, dir, "jj", "git", "init")
	runCmd(t, dir, "jj", "bookmark", "create", "main", "-r", "@")

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	result, ok := generated.(JjResult)
	require.True(t, ok)
	require.Equal(t, "main", result.Branch)
	require.Equal(t, "main", result.Bookmark)
	require.Empty(t, result.LastBookmark, "no bookmarked ancestor above @ when the bookmark is at @ itself")
}

func TestJjOperationReportsDirtyWorkingCopy(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()
	runCmd(t, dir, "jj", "git", "init")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hello"), 0644))

	op := &Jj{}
	generated, err := op.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	result, ok := generated.(JjResult)
	require.True(t, ok)
	require.Contains(t, result.Status, "dirty")
}

func TestJjAndGitOperationsCanCoexistInOneTemplate(t *testing.T) {
	requireBinary(t, "jj")
	dir := t.TempDir()
	runCmd(t, dir, "jj", "git", "init")
	runCmd(t, dir, "jj", "bookmark", "create", "main", "-r", "@")

	jjOp := &Jj{}
	jjResult, err := jjOp.Generate("pane", "tmux.%1", dir, "")
	require.NoError(t, err)

	// Both operations always run (Generate doesn't know or care what the
	// template does with its output) — it's the config template that picks
	// jj over git when jj is available, e.g.
	// `{{ if .jj }}...{{ else if .git }}...{{ end }}`. gitResult is left nil
	// here (as it would be for an uncommitted jj repo with no exported git
	// commit yet) specifically to prove the template's jj-branch doesn't
	// depend on git having anything to say at all.
	var gitResult interface{}

	tmpl, err := template.New("t").Parse(`{{- if .jj }}jj:{{ .jj.Branch }}{{ else if .git }}git:{{ .git.Branch }}{{ end -}}`)
	require.NoError(t, err)

	var buf strings.Builder
	require.NoError(t, tmpl.Execute(&buf, map[string]interface{}{"jj": jjResult, "git": gitResult}))
	require.Equal(t, "jj:main", buf.String())
}

func TestGitTemplateFallbackWhenJjIsNil(t *testing.T) {
	tmpl, err := template.New("t").Parse(`{{- if .jj }}jj:{{ .jj.Branch }}{{ else if .git }}git:{{ .git.Branch }}{{ end -}}`)
	require.NoError(t, err)

	var buf strings.Builder
	data := map[string]interface{}{"jj": nil, "git": GitResult{Branch: "main"}}
	require.NoError(t, tmpl.Execute(&buf, data))
	require.Equal(t, "git:main", buf.String())
}
