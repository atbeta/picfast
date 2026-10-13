package service

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	StartupImaging()
	code := m.Run()
	ShutdownImaging()
	os.Exit(code)
}
