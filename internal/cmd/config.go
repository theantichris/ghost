package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
)

// loadConfigFile attempts to load config.toml from ~/.config/ghost and returns a StringPtrSourcer.
// If the config file does not exist, it returns an empty sourcer without error allowing the application to use default flag values. Returns ErrConfigFile if the home directory cannot be determined.
func loadConfigFile(logger *log.Logger) (altsrc.StringPtrSourcer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return altsrc.StringPtrSourcer{}, fmt.Errorf("%w", ErrConfigFile)
	}

	configFile := filepath.Join(homeDir, ".config/ghost", "config.toml")

	var sourcer altsrc.StringPtrSourcer
	if _, err := os.Stat(configFile); err != nil {
		logger.Debug("config file not found", "file", configFile)
	} else {
		sourcer = altsrc.NewStringPtrSourcer(&configFile)
		logger.Debug("loading config file", "file", configFile)
	}

	return sourcer, nil
}
