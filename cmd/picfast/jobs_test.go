package main

import (
	"bytes"
	"context"
	"testing"
)

func TestRunJobsCommandUsage(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"help", []string{"help"}, 0},
		{"short help", []string{"-h"}, 0},
		{"long help", []string{"--help"}, 0},
		{"no args", []string{}, 2},
		{"unknown command", []string{"bogus"}, 2},
		{"extra argument", []string{"run", "extra"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runJobsCommand(context.Background(), tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("runJobsCommand(%v) = %d, want %d (stdout=%q stderr=%q)", tt.args, code, tt.wantCode, stdout.String(), stderr.String())
			}
		})
	}
}
