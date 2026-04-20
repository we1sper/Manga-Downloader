package script

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed script.js
var script string

func Get() string {
	return script
}

func LoadScript(location string) error {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return fmt.Errorf("failed to load script: %v", err)
	}
	script = string(bytes)
	return nil
}
