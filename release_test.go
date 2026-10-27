package simplyddns

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseVersionMetadata(t *testing.T) {
	for _, tc := range []struct {
		input, version, prerelease string
	}{
		{"2.0.0", "2.0.0", "false"},
		{"v2.0.0", "2.0.0", "false"},
		{"2.0.0-rc.1", "2.0.0-rc.1", "true"},
		{"2.0.0-0+build.1", "2.0.0-0+build.1", "true"},
		{"2.0.0+build-with-hyphens", "2.0.0+build-with-hyphens", "false"},
		{"", "", ""}, {"2.0", "", ""}, {"v", "", ""}, {"02.0.0", "", ""},
		{"2.0.0-01", "", ""}, {"2.0.0-rc..1", "", ""}, {"2.0.0+", "", ""},
		{"2.0.0; exit 0", "", ""},
		{"2.0.0-\u03b1", "", ""},
	} {
		t.Run(tc.input, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", "scripts/release-version.sh", tc.input)
			output, err := cmd.CombinedOutput()
			if tc.version == "" {
				require.Error(t, err, string(output))
				assert.Contains(t, string(output), "Invalid release version")
				return
			}
			require.NoError(t, err, string(output))
			assert.Equal(t, fmt.Sprintf("version=%s\ntag=v%s\nprerelease=%s\n", tc.version, tc.version, tc.prerelease), string(output))
		})
	}
}

func TestReleaseArtifacts(t *testing.T) {
	dir := os.Getenv("SIMPLYDDNS_DIST_DIR")
	if dir == "" {
		t.Skip("run mise exec -- make verify-dist to build and verify release archives")
	}
	version := os.Getenv("SIMPLYDDNS_VERSION")
	bin := os.Getenv("SIMPLYDDNS_BIN")
	targets := strings.Fields(os.Getenv("SIMPLYDDNS_TARGETS"))
	require.NotEmpty(t, version)
	require.NotEmpty(t, bin)
	require.NotEmpty(t, targets)
	checksums, err := os.ReadFile(filepath.Join(dir, "checksums.txt"))
	require.NoError(t, err)
	hashes := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(string(checksums)), "\n") {
		fields := strings.Fields(line)
		require.Len(t, fields, 2)
		require.Len(t, fields[0], 64)
		require.NotContains(t, hashes, fields[1], "duplicate checksum entry")
		hashes[fields[1]] = fields[0]
	}
	require.Len(t, hashes, len(targets))
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			goos, goarch, ok := strings.Cut(target, "/")
			require.True(t, ok)
			binaryName := bin
			extension := ".tar.gz"
			if goos == "windows" {
				binaryName += ".exe"
				extension = ".zip"
			}
			name := fmt.Sprintf("%s_%s_%s_%s%s", bin, version, goos, goarch, extension)
			data, err := os.ReadFile(filepath.Join(dir, name))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(data)), hashes[name], "archive checksum mismatch")
			files := readReleaseArchive(t, data, extension)
			expected := []string{binaryName, "LICENSE", "README.md", "README_ZH.md", "simplyddns.yml"}
			require.Len(t, files, len(expected))
			for _, filename := range expected {
				require.NotEmpty(t, files[filename], "%s is missing or empty", filename)
			}
			info, err := buildinfo.Read(bytes.NewReader(files[binaryName]))
			require.NoError(t, err)
			assert.Equal(t, "github.com/mingcheng/simplyddns/cmd/simplyddns", info.Path)
			assert.Equal(t, runtime.Version(), info.GoVersion)
			settings := make(map[string]string)
			for _, setting := range info.Settings {
				settings[setting.Key] = setting.Value
			}
			assert.Equal(t, goos, settings["GOOS"])
			assert.Equal(t, goarch, settings["GOARCH"])
			assert.Equal(t, "0", settings["CGO_ENABLED"])
			if goos == runtime.GOOS && goarch == runtime.GOARCH {
				temp := t.TempDir()
				path := filepath.Join(temp, binaryName)
				require.NoError(t, os.WriteFile(path, files[binaryName], 0700))
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, path, "--version")
				cmd.Dir = temp
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, string(output))
				assert.Contains(t, string(output), "simplyddns "+version+" (commit ")
				assert.NotContains(t, string(output), "unknown")
			}
		})
	}
}

func readReleaseArchive(t *testing.T, data []byte, extension string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	add := func(name string, reader io.Reader) {
		t.Helper()
		require.Equal(t, filepath.Base(name), name, "archive entries must not escape the extraction directory")
		require.NotContains(t, files, name)
		body, err := io.ReadAll(reader)
		require.NoError(t, err)
		files[name] = body
	}
	if extension == ".zip" {
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		require.NoError(t, err)
		for _, file := range reader.File {
			require.True(t, file.Mode().IsRegular())
			body, err := file.Open()
			require.NoError(t, err)
			add(file.Name, body)
			require.NoError(t, body.Close())
		}
		return files
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Equal(t, byte(tar.TypeReg), header.Typeflag)
		if header.Name == os.Getenv("SIMPLYDDNS_BIN") {
			assert.NotZero(t, header.Mode&0111, "Unix binaries must be executable")
		}
		add(header.Name, reader)
	}
	_, err = io.Copy(io.Discard, compressed)
	require.NoError(t, err)
	return files
}
