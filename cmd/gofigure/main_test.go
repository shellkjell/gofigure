package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fig")
	if err := os.WriteFile(testFile, []byte(`key: "value"`), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:     "no input file",
			args:     []string{},
			wantCode: exitUsage,
		},
		{
			name:       "parse simple file",
			args:       []string{"-i", testFile},
			wantCode:   exitSuccess,
			wantStdout: `"key"`,
		},
		{
			name:       "parse with indent",
			args:       []string{"-i", testFile, "-indent"},
			wantCode:   exitSuccess,
			wantStdout: "{\n",
		},
		{
			name:       "version flag",
			args:       []string{"-version"},
			wantCode:   exitSuccess,
			wantStdout: "gofigure",
		},
		{
			name:       "nonexistent file",
			args:       []string{"-i", "/nonexistent/file.fig"},
			wantCode:   exitError,
			wantStderr: "Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			code := run(tc.args, stdout, stderr)

			if code != tc.wantCode {
				t.Errorf("run() = %d, want %d\nstderr: %s", code, tc.wantCode, stderr.String())
			}

			if tc.wantStdout != "" && !strings.Contains(stdout.String(), tc.wantStdout) {
				t.Errorf("stdout = %q, want to contain %q", stdout.String(), tc.wantStdout)
			}

			if tc.wantStderr != "" && !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("stderr = %q, want to contain %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}

func TestRunOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.fig")
	outputFile := filepath.Join(tempDir, "output.json")

	if err := os.WriteFile(testFile, []byte(`key: "value"`), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := run([]string{"-i", testFile, "-o", outputFile}, stdout, stderr)
	if code != exitSuccess {
		t.Fatalf("run() = %d, want %d\nstderr: %s", code, exitSuccess, stderr.String())
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "key") {
		t.Errorf("output file content = %q, want to contain 'key'", string(content))
	}
}
