package recovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanePath(t *testing.T) {
	t.Setenv("GH_DASH_STATE_DIR", t.TempDir())
	t.Setenv("REX_BLOCK", "")
	path, err := PanePath()
	require.NoError(t, err)
	require.Empty(t, path)
	t.Setenv("REX_BLOCK", "block:first")
	first, err := PanePath()
	require.NoError(t, err)
	t.Setenv("REX_BLOCK", "block:second")
	second, err := PanePath()
	require.NoError(t, err)
	require.NotEqual(t, first, second)
	t.Setenv("REX_BLOCK", "../../escape")
	path, err = PanePath()
	require.NoError(t, err)
	require.Equal(t, os.Getenv("GH_DASH_STATE_DIR"), filepath.Dir(path))
}

func TestStore_RoundTripLockAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pane", "state.json")
	store, err := Open(path)
	require.NoError(t, err)
	c := Checkpoint{Active: true, Cwd: t.TempDir(), UI: json.RawMessage(`{"selected":"https://github.com/org/repo/pull/42"}`)}
	require.NoError(t, store.Save(c))
	saved, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, Version, saved.Version)
	require.True(t, saved.Active)
	require.JSONEq(t, string(c.UI), string(saved.UI))
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Nested shells and concurrent launches cannot steal the checkpoint.
	_, err = Open(path)
	require.ErrorIs(t, err, ErrBusy)
	require.NoError(t, store.Save(c))
	unchanged, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, info.ModTime(), unchanged.ModTime(), "unchanged frames must not write")

	c.Active = false
	require.NoError(t, store.Save(c))
	require.NoError(t, store.Close())
	store, err = Open(path)
	require.NoError(t, err)
	defer store.Close()
	saved, err = Load(path)
	require.NoError(t, err)
	require.False(t, saved.Active)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".checkpoint-*"))
	require.NoError(t, err)
	require.Empty(t, files)
}

func TestLoad_RejectsBadCheckpoints(t *testing.T) {
	for name, content := range map[string]string{
		"truncated":       `{"version":1`,
		"future version":  `{"version":99,"cwd":"/repo","ui":{}}`,
		"relative cwd":    `{"version":1,"cwd":"relative","ui":{}}`,
		"relative config": `{"version":1,"cwd":"/repo","config":"config.yml","ui":{}}`,
		"oversized":       strings.Repeat(" ", maxCheckpointSize+1),
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "state.json")
			require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
			_, err := Load(path)
			require.Error(t, err)
		})
	}
}
