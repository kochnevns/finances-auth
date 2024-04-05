package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)
func TestMustLoad(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tempFile := tempfile()
	defer os.Remove(tempFile)

	os.Args = append(os.Args, "-config", tempFile)

	err := ioutil.WriteFile(tempFile, []byte(`
env: local
storage_path: test-storage
grpc:
  port: 1234
  timeout: 1s
token_ttl: 1h
`), 0666)

	if err != nil {
		t.Fatal(err)
	}

	cfg := MustLoad()

	if cfg.Env != "local" {
		t.Errorf("wrong env, expected local, got %s", cfg.Env)
	}

	if cfg.StoragePath != "test-storage" {
		t.Errorf("wrong storage path, expected test-storage, got %s", cfg.StoragePath)
	}

	if cfg.GRPC.Port != 1234 {
		t.Errorf("wrong grpc port, expected 1234, got %d", cfg.GRPC.Port)
	}

	if cfg.GRPC.Timeout != 1*time.Second {
		t.Errorf("wrong grpc timeout, expected 1s, got %v", cfg.GRPC.Timeout)
	}

	if cfg.TokenTTL != time.Hour {
		t.Errorf("wrong token ttl, expected 1h, got %v", cfg.TokenTTL)
	}
}

func tempfile() string {
	f, err := ioutil.TempFile("", "test-*.yaml")
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}

	return f.Name()
}
