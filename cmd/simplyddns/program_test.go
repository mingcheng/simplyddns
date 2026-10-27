package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgramStartStop(t *testing.T) {
	for range 20 {
		program := &Program{Configure: &simplyddns.Config{Jobs: []simplyddns.JobConfig{{
			Source: simplyddns.SourceConfig{Type: "lo"},
			Target: simplyddns.TargetConfig{Type: "sleep"},
		}}}}
		require.NoError(t, program.Start())
		require.NoError(t, program.Stop())
		require.NoError(t, program.Stop())
	}
	require.Error(t, (&Program{}).Start())
	require.Error(t, (&Program{Configure: &simplyddns.Config{}}).Start())
}

func TestProgramClosesLogFile(t *testing.T) {
	oldOutput, oldLevel := log.Out, log.GetLevel()
	t.Cleanup(func() {
		log.SetOutput(oldOutput)
		log.SetLevel(oldLevel)
	})
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yml")
	logFile := filepath.Join(dir, "ddns.log")
	config := "logfile: " + filepath.ToSlash(logFile) + "\ndebug: true\nddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"
	require.NoError(t, os.WriteFile(configFile, []byte(config), 0600))
	program := &Program{Configure: &simplyddns.Config{}, ConfigFile: configFile}
	require.NoError(t, program.Init(nil))
	fp := program.logFile
	require.NotNil(t, fp)
	require.NoError(t, program.Start())
	require.NoError(t, program.Stop())
	_, err := fp.WriteString("closed")
	assert.ErrorIs(t, err, os.ErrClosed)
	require.NoError(t, program.Stop())
	data, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestProgramUsesStandardStreams(t *testing.T) {
	oldOutput, oldLevel := log.Out, log.GetLevel()
	t.Cleanup(func() {
		log.SetOutput(oldOutput)
		log.SetLevel(oldLevel)
	})
	for _, tc := range []struct {
		path string
		file *os.File
	}{
		{"/dev/stdout", os.Stdout},
		{"/dev/stderr", os.Stderr},
	} {
		t.Run(tc.path, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "config.yml")
			config := "logfile: " + tc.path + "\nddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"
			require.NoError(t, os.WriteFile(filename, []byte(config), 0600))
			program := &Program{Configure: &simplyddns.Config{}, ConfigFile: filename}
			require.NoError(t, program.Init(nil))
			assert.Same(t, tc.file, log.Out)
			assert.Nil(t, program.logFile, "standard streams must not be owned or closed by the program")
			require.NoError(t, program.Stop())
		})
	}
}
