/*!*
 * Copyright (c) 2025 Hangzhou Guanwaii Technology Co,.Ltd.
 *
 * This source code is licensed under the MIT License,
 * which is located in the LICENSE file in the source tree's root directory.
 *
 * File: config_test.go
 * Author: mingcheng@outlook.com
 * File Created: Saturday, December 26th 2020
 *
 * Modified By: mingcheng (mingcheng@apache.org)
 * Last Modified: 2025-03-12 14:36:30
 */

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/mingcheng/simplyddns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mingcheng/simplyddns/target"
)

func TestReadConfigure(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.yml")
	require.NoError(t, os.WriteFile(filename, []byte(`
debug: true
defaults: &defaults
  type: lo,static
  content: "127.0.0.1"
  interval: 60
ddns:
  - source:
      <<: *defaults
    target:
      type: sleep
    webhook:
      url: https://example.com/hook
      token: test-token
`), 0600))
	var config simplyddns.Config
	require.NoError(t, readConfigure(&config, filename))
	require.Len(t, config.Jobs, 1)
	assert.True(t, config.Debug)
	assert.Equal(t, "lo,static", config.Jobs[0].Source.Type)
	assert.Equal(t, "test-token", config.Jobs[0].WebHook.Token)
	job, err := simplyddns.NewJob(config.Jobs[0])
	require.NoError(t, err)
	job.Stop()

	require.NoError(t, os.WriteFile(filename, []byte("ddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"), 0600))
	require.NoError(t, readConfigure(&config, filename))
	assert.False(t, config.Debug, "configuration reloads must not retain old values")
	assert.Empty(t, config.Jobs[0].WebHook.Url)
}

func TestReadConfigureRejectsInvalidInput(t *testing.T) {
	for _, content := range []string{
		"invalid: [", "{}", "ddns: []",
		"ddns:\n  - source: {type: lo, interval: -1}\n    target: {type: sleep}",
		"ddns:\n  - source: {type: lo}\n    target: {type: sleep, domains: [https://example.com/path]}",
	} {
		t.Run(content, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(filename, []byte(content), 0600))
			require.Error(t, readConfigure(&simplyddns.Config{}, filename))
		})
	}
	require.Error(t, ReadConfigure(nil))
	require.Error(t, readConfigure(&simplyddns.Config{}, filepath.Join(t.TempDir(), "missing.yml")))
}

func TestDefaultConfigurationName(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "simplyddns.yml"), []byte("ddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"), 0600))
	var config simplyddns.Config
	require.NoError(t, readConfigure(&config, "", dir))
	require.Len(t, config.Jobs, 1)
}

func TestExampleConfigurations(t *testing.T) {
	files, err := filepath.Glob("../../example/*.yml")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, filename := range files {
		t.Run(filepath.Base(filename), func(t *testing.T) {
			var config simplyddns.Config
			require.NoError(t, readConfigure(&config, filename))
			dispatch, err := simplyddns.NewDispatch(config.Jobs)
			require.NoError(t, err)
			dispatch.Stop()
		})
	}
}

func TestCommandFlags(t *testing.T) {
	var output bytes.Buffer
	require.NoError(t, run([]string{"--version", "--config", "does-not-exist.yml"}, &output))
	assert.Contains(t, output.String(), AppName+" "+BuildVersion)
	assert.Contains(t, output.String(), BuildCommit)
	output.Reset()
	require.NoError(t, run([]string{"--help"}, &output))
	assert.Contains(t, output.String(), "-config")
	require.Error(t, run([]string{"--unknown"}, &output))
	require.Error(t, run([]string{"unexpected"}, &output))
}

func TestConfigurationSearchOrder(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(first, "simplyddns.yaml"), []byte("debug: true\nddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(second, "simplyddns.yml"), []byte("debug: false\nddns:\n  - source: {type: lo}\n    target: {type: sleep}\n"), 0600))
	var config simplyddns.Config
	require.NoError(t, readConfigure(&config, "", first, second))
	assert.True(t, config.Debug)
}
