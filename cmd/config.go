package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrNoModel          = errors.New("model is required (set via --model flag, config file, or environment)")
	ErrNoVisionModel    = errors.New("vision model is required with images (set via --vision-model flag, config file, or environment)")
	ErrInvalidFormat    = errors.New("invalid format option, valid options are json or markdown")
	ErrInvalidImageFlag = errors.New("failed to read image flag")
	ErrConfig           = errors.New("failed to read config file")
	ErrBindFlags        = errors.New("failed to bind flags")
)

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command, cfgFile string) error {
	viper.SetEnvPrefix("GHOST")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(filepath.Join(home, ".config", "ghost"))
		viper.SetConfigName("config.toml")
		viper.SetConfigType("toml")
	}

	logger := cmd.Context().Value(loggerKey{}).(*log.Logger)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("%w: %w", ErrConfig, err)
		}

		logger.Debug("no config file found, using flags/env only")
	} else {
		logger.Debug("loaded config", "file", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrBindFlags, err)
	}

	model := viper.GetString("model")
	if model == "" {
		return ErrNoModel
	}

	err = validateFormat(viper.GetString("format"))
	if err != nil {
		return err
	}

	_ = viper.BindPFlag("vision.model", cmd.Flags().Lookup("vision-model"))

	imagePaths, err := cmd.Flags().GetStringArray("image")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidImageFlag, err)
	}

	if len(imagePaths) > 0 && viper.GetString("vision.model") == "" {
		return ErrNoVisionModel
	}

	return nil
}

// validateFormat returns an error if the format flag isn't a valid value.
func validateFormat(format string) error {
	if format != "" && (format != "json" && format != "markdown") {
		return ErrInvalidFormat
	}

	return nil
}
