package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Retry          uint32
	Cookie         string
	SaveDir        string
	Proxy          string
	ScriptLocation string
	Concurrency    uint32
	Timeout        uint32 // Milliseconds.
	Overwrite      bool
}

func Default() *Config {
	return &Config{
		Retry:       5,
		Concurrency: 3,
		Timeout:     60000,
	}
}

func (cfg *Config) Load(path string) error {
	bytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return readErr
	}
	unmarshalErr := json.Unmarshal(bytes, cfg)
	if unmarshalErr != nil {
		return unmarshalErr
	}
	return nil
}
