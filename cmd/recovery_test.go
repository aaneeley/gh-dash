package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/recovery"
	"github.com/stretchr/testify/require"
)

func TestPrepareRecovery_ResumeIsSilentWithoutActiveState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pane.json")
	store, _, err := prepareRecovery(path, true)
	require.NoError(t, err)
	require.Nil(t, store)

	s, err := recovery.Open(path)
	require.NoError(t, err)
	require.NoError(t, s.Save(recovery.Checkpoint{Cwd: t.TempDir(), Active: false, UI: json.RawMessage(`{}`)}))
	s.Close()
	store, _, err = prepareRecovery(path, true)
	require.NoError(t, err)
	require.Nil(t, store)
}

func TestPrepareRecovery_RestoresCwdAndConfigWithoutRunningCommands(t *testing.T) {
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	oldConfig := cfgFlag
	t.Cleanup(func() { _ = os.Chdir(oldCwd); cfgFlag = oldConfig })
	path := filepath.Join(t.TempDir(), "pane.json")
	cwd := t.TempDir()
	cfg := filepath.Join(cwd, "config with spaces;literal.yml")
	require.NoError(t, os.WriteFile(cfg, []byte("prSections: []\n"), 0o600))
	s, err := recovery.Open(path)
	require.NoError(t, err)
	require.NoError(t, s.Save(recovery.Checkpoint{Cwd: cwd, Config: cfg, Active: true, UI: json.RawMessage(`{}`)}))
	s.Close()
	store, saved, err := prepareRecovery(path, true)
	require.NoError(t, err)
	require.NotNil(t, store)
	defer store.Close()
	require.Equal(t, cfg, cfgFlag)
	actual, err := os.Getwd()
	require.NoError(t, err)
	expected, err := filepath.EvalSymlinks(cwd)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
	require.True(t, saved.Active)

	// Starting an interactive nested shell while Dash is running is a no-op.
	busy, _, err := prepareRecovery(path, true)
	require.NoError(t, err)
	require.Nil(t, busy)
}
