package config

import (
	"fmt"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	cfg := Default()
	if err := cfg.Load("./sample/config.json"); err != nil {
		t.Fatalf("load config error: %v", err)
	}
	fmt.Println(cfg)
}
