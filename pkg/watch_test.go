package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// startTmuxTestServer boots a throwaway tmux server on its own socket (so
// tests never touch the user's real one) with the session status line off
// and the pane border driven by borderFormat, mirroring the pane-only setup
// tmux.conf uses.
func startTmuxTestServer(t *testing.T, socket, borderFormat string) {
	t.Helper()
	requireBinary(t, "tmux")

	run := func(args ...string) {
		out, err := exec.Command("tmux", append([]string{"-L", socket}, args...)...).CombinedOutput()
		require.NoErrorf(t, err, "tmux %s: %s", strings.Join(args, " "), out)
	}
	run("new-session", "-d", "-x", "100", "-y", "20", "sleep", "60")
	t.Cleanup(func() {
		exec.Command("tmux", "-L", socket, "kill-server").Run()
	})
	run("set-option", "-g", "status", "off")
	run("set-option", "-g", "pane-border-status", "bottom")
	run("set-option", "-g", "pane-border-format", borderFormat)
}

// attachTmuxClient attaches a throwaway client through script's pty. Without
// an attached client nothing renders and pane-border-format's #(...) jobs
// never run.
func attachTmuxClient(t *testing.T, socket string) {
	t.Helper()
	requireBinary(t, "script")

	cmd := exec.Command("script", "-q", "/dev/null", "tmux", "-L", socket, "attach", "-d")
	cmd.Env = append(envWithoutTMUX(), "TERM=xterm-256color")
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
	})

	deadline := time.Now().Add(5 * time.Second)
	for {
		out, err := exec.Command("tmux", "-L", socket, "list-clients", "-F", "#{client_name}").CombinedOutput()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("client never attached: %s", out)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func envWithoutTMUX() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "TMUX=") {
			env = append(env, kv)
		}
	}
	return env
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return len(strings.Split(strings.TrimSpace(string(data)), "\n"))
}

// waitForHits polls until the border command has run more than want times,
// since #() jobs run asynchronously after the redraw.
func waitForHits(t *testing.T, hits string, want int) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		got := countLines(t, hits)
		if got > want {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("border command ran %d times, want more than %d", got, want)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestRefreshTmuxClientsReRunsPaneBorderCommands(t *testing.T) {
	requireBinary(t, "tmux")
	requireBinary(t, "script")
	dir := t.TempDir()
	hits := filepath.Join(dir, "hits")

	counter := filepath.Join(dir, "counter.sh")
	require.NoError(t, os.WriteFile(counter, []byte("#!/bin/sh\necho 1 >> "+hits+"\n"), 0755))

	socket := fmt.Sprintf("clt_test_%d", os.Getpid())
	startTmuxTestServer(t, socket, "#("+counter+")")
	attachTmuxClient(t, socket)

	before := waitForHits(t, hits, 0)
	require.NoError(t, RefreshTmuxClients(socket))
	waitForHits(t, hits, before)
}

func TestWatchExitsWhenWatcherAlreadyRunning(t *testing.T) {
	requireBinary(t, "tmux")
	socket := fmt.Sprintf("clt_test_%d", os.Getpid())
	startTmuxTestServer(t, socket, "#{pane_id}")

	logger := log.New(io.Discard, "", 0)

	// A fresh server has no recorded watcher: we may claim it.
	ok, err := EnsureSingleWatcher(socket)
	require.NoError(t, err)
	require.True(t, ok)

	// Our own pid is now recorded and alive: claiming again must fail.
	ok, err = EnsureSingleWatcher(socket)
	require.NoError(t, err)
	require.False(t, ok)

	// And Watch must exit immediately rather than loop forever.
	require.NoError(t, Watch(time.Hour, socket, logger))
}

func TestWatchRefreshesPeriodically(t *testing.T) {
	requireBinary(t, "tmux")
	requireBinary(t, "script")
	dir := t.TempDir()
	hits := filepath.Join(dir, "hits")

	counter := filepath.Join(dir, "counter.sh")
	require.NoError(t, os.WriteFile(counter, []byte("#!/bin/sh\necho 1 >> "+hits+"\n"), 0755))

	socket := fmt.Sprintf("clt_test_%d", os.Getpid())
	startTmuxTestServer(t, socket, "#("+counter+")")
	attachTmuxClient(t, socket)

	done := make(chan error, 1)
	go func() { done <- Watch(100*time.Millisecond, socket, log.New(io.Discard, "", 0)) }()
	waitForHits(t, hits, 2)

	// Watch keeps looping until the server stops answering, so end it
	// explicitly (the registered cleanup is idempotent) and make sure the
	// goroutine unwinds rather than leaking past the test.
	exec.Command("tmux", "-L", socket, "kill-server").Run()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Watch did not exit after the server died")
	}
}
