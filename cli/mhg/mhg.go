package main

import (
	"fmt"

	"github.com/we1sper/Manga-Downloader/mhg"
	"github.com/we1sper/Manga-Downloader/mhg/config"
	"github.com/we1sper/Manga-Downloader/pkg/command"
	"github.com/we1sper/Manga-Downloader/pkg/util"
)

var (
	configPath = "./config.json"
	cmd        = command.InitializeCommand()
)

func init() {
	generateConfigArgument := command.NewMarkArgument("template", "t", "generate a template config file under current work directory").Action(func() error {
		template := &config.Config{
			ApiServerPort: 8080,
			Retry:         5,
			SaveDir:       "path to save, e.g., ./downloads",
			Proxy:         "http proxy address, e.g., http://localhost:10808",
			Concurrency:   3,
			Timeout:       60000,
			Overwrite:     false,
		}
		if err := util.SaveToJsonFile(configPath, template); err != nil {
			return fmt.Errorf("failed to generate template config file: %w", err)
		}
		return nil
	})

	specifyConfigArgument := command.NewValueArgument("config", "c", "specify config file path, the default path is './config.json'").Action(func(values []string) error {
		if len(values) > 0 {
			configPath = values[0]
		}
		return nil
	})

	serveModeArgument := command.NewMarkArgument("serve", "s", "enable serve mode: submit download tasks through http requests").Action(func() error {
		cfg := config.Default()
		if err := cfg.Load(configPath); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		server, err := mhg.NewApiServer(cfg)
		if err != nil {
			return fmt.Errorf("failed to create api server: %w", err)
		}

		server.Run()

		return nil
	})

	cmd.Register(generateConfigArgument).Register(specifyConfigArgument).Register(serveModeArgument).EnableHelp()
}

func main() {
	cmd.Pipeline()
}
