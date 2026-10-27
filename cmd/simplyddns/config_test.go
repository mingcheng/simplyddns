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
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestReadConfigureAutomaticSources(t *testing.T) {
	t.Setenv(simplyddns.AutoSourceCountEnv, "3")
	for _, tc := range []struct {
		source string
		want   string
	}{
		{"", "auto"},
		{"{}", "auto"},
		{"{interval: 300}", "auto"},
		{"{type: ''}", "auto"},
		{"{type: '   '}", "auto"},
		{"{type: auto}", "auto"},
		{"{type: auto-ipv4}", "auto-ipv4"},
		{"{type: auto-ipv6}", "auto-ipv6"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			content := "ddns:\n  - target: {type: sleep}\n"
			if tc.source != "" {
				content += "    source: " + tc.source + "\n"
			}
			filename := filepath.Join(t.TempDir(), "config.yml")
			require.NoError(t, os.WriteFile(filename, []byte(content), 0600))
			var config simplyddns.Config
			require.NoError(t, readConfigure(&config, filename))
			require.Len(t, config.Jobs, 1)
			job, err := simplyddns.NewJob(config.Jobs[0])
			require.NoError(t, err)
			defer job.Stop()
			assert.Equal(t, tc.want, job.Config.Source.Type)
		})
	}
}

func TestReadConfigureRejectsInvalidAutomaticCount(t *testing.T) {
	for _, source := range []string{"{}", "{type: auto}", "{type: auto-ipv4}", "{type: auto-ipv6}"} {
		for _, count := range []string{"", "4", "9", "invalid"} {
			t.Run(source+"/"+count, func(t *testing.T) {
				t.Setenv(simplyddns.AutoSourceCountEnv, count)
				filename := filepath.Join(t.TempDir(), "config.yml")
				content := fmt.Sprintf("ddns:\n  - source: %s\n    target: {type: sleep}\n", source)
				require.NoError(t, os.WriteFile(filename, []byte(content), 0600))
				err := readConfigure(&simplyddns.Config{}, filename)
				require.ErrorContains(t, err, "job 1")
				assert.ErrorContains(t, err, simplyddns.AutoSourceCountEnv)
			})
		}
	}
}

func TestReadConfigureRejectsInvalidInput(t *testing.T) {
	for _, content := range []string{
		"invalid: [", "{}", "ddns: []",
		"ddns:\n  - source: {type: lo, interval: -1}\n    target: {type: sleep}",
		"ddns:\n  - source: {type: lo}\n    target: {type: sleep, domains: [https://example.com/path]}",
		"ddns:\n  - source: {type: 'auto,ipify'}\n    target: {type: sleep}",
		"ddns:\n  - source: {type: 'auto-ipv6,auto-ipv4'}\n    target: {type: sleep}",
		"ddns:\n  - source: {type: 'ipify,'}\n    target: {type: sleep}",
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
	t.Setenv(simplyddns.AutoSourceCountEnv, "3")
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

func TestReadmeConfigurations(t *testing.T) {
	t.Setenv(simplyddns.AutoSourceCountEnv, "3")
	for _, name := range []string{"README.md", "README_ZH.md"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../..", name))
			require.NoError(t, err)
			_, rest, ok := strings.Cut(string(data), "```yaml")
			require.True(t, ok, "missing quick-start configuration")
			example, _, ok := strings.Cut(rest, "```")
			require.True(t, ok, "unclosed configuration example")

			filename := filepath.Join(t.TempDir(), "simplyddns.yml")
			require.NoError(t, os.WriteFile(filename, []byte(example), 0600))
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
