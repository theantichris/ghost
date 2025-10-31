package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
)

// loadConfigFile attempts to load config.toml from ~/.config/ghost and returns a StringSourcer.
// If the config file does not exist, it returns an empty string without error allowing the application to use default flag values. Returns ErrConfigFile if the home directory cannot be determined.
func loadConfigFile(logger *log.Logger) (altsrc.StringSourcer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w", ErrConfigFile)
	}

	configFile := filepath.Join(homeDir, ".config/ghost", "config.toml")

	if _, err := os.Stat(configFile); err != nil {
		logger.Debug("no config file to load", "file", configFile)
	} else {
		logger.Debug("loading config file", "file", configFile)
	}

	return altsrc.StringSourcer(configFile), nil
}
