package pkg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// tmuxArgs prefixes a tmux invocation with the -L socket flag when a
// non-default socket is given (used by tests to run against a throwaway
// server instead of the user's real one).
func tmuxArgs(socketName string, args ...string) []string {
	if socketName != "" {
		return append([]string{"-L", socketName}, args...)
	}
	return args
}

func runTmux(socketName string, args ...string) (string, error) {
	out, err := exec.Command("tmux", tmuxArgs(socketName, args...)...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// WatcherPIDOption is the tmux user option Watch stores its own pid in so a
// second Watch (e.g. a config reload's run-shell starting another one) can
// tell the first is still alive and exit instead of stacking loops.
const WatcherPIDOption = "@commandline_thing_watcher_pid"

// EnsureSingleWatcher returns false when a watcher is already running for
// this tmux server, otherwise records our pid in WatcherPIDOption and
// returns true. A stale pid (a watcher that died without cleanup) is
// overwritten.
func EnsureSingleWatcher(socketName string) (bool, error) {
	if pidStr, err := runTmux(socketName, "show-options", "-gv", WatcherPIDOption); err == nil && pidStr != "" {
		var pid int
		if _, err := fmt.Sscanf(pidStr, "%d", &pid); err == nil && syscall.Kill(pid, 0) == nil {
			return false, nil
		}
	}

	_, err := runTmux(socketName, "set-option", "-gq", WatcherPIDOption, fmt.Sprint(os.Getpid()))
	return err == nil, err
}

// RefreshTmuxClients asks every attached client to redraw its status and
// pane borders. tmux only re-evaluates pane-border-format's #(...) commands
// when the border is redrawn, and with the session status line off it never
// redraws borders on its own, so an explicit refresh is what keeps
// operations that shell out at Generate time (jj, git) fresh while panes
// are idle.
func RefreshTmuxClients(socketName string) error {
	clients, err := runTmux(socketName, "list-clients", "-F", "#{client_name}")
	if err != nil {
		return err
	}
	for _, client := range strings.Split(clients, "\n") {
		if strings.TrimSpace(client) == "" {
			continue
		}
		runTmux(socketName, "refresh-client", "-S", "-t", client)
	}
	return nil
}

// Watch refreshes clients immediately and then every interval, returning
// once the tmux server stops answering (a watcher started via run-shell is
// a child of the server and dies with it anyway). Should be started once
// per server, e.g. from tmux.conf via run-shell -b.
func Watch(interval time.Duration, socketName string, logger *log.Logger) error {
	ok, err := EnsureSingleWatcher(socketName)
	if err != nil {
		return fmt.Errorf("checking for a running watcher: %w", err)
	}
	if !ok {
		logger.Printf("watcher already running, exiting")
		return nil
	}

	if err := RefreshTmuxClients(socketName); err != nil {
		logger.Printf("initial refresh failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if err := RefreshTmuxClients(socketName); err != nil {
			logger.Printf("stopping: %v", err)
			return err
		}
	}
	return nil
}
