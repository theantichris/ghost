package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Variable to store flags and Logger.
var (
	configFile string
	Debug      bool
	Ollama     string
	Model      string
	Logger     *slog.Logger
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "ghost",
	Short: "A cyberpunk inspired AI assistant.",
	Long:  "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		Logger.Error("error running ghost command", slog.String("component", "cmd.RootCmd"))
		os.Exit(1)
	}
}

// init initializes the application.
func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.ghost.toml)")
	RootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug mode")
	RootCmd.PersistentFlags().StringVar(&Model, "model", "", "LLM model to use")
	RootCmd.PersistentFlags().StringVar(&Ollama, "ollama", "", "Ollama API base URL")

	_ = viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("ollama", RootCmd.PersistentFlags().Lookup("ollama"))
	_ = viper.BindPFlag("model", RootCmd.PersistentFlags().Lookup("model"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initLogger()

	if err := godotenv.Load(); err != nil {
		Logger.Debug(".env file not found, using environment variables", "component", "cmd.RootCmd")
	} else {
		Logger.Debug(".env file loaded successfully", "component", "cmd.RootCmd")
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in home directory and current directory.
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ghost")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()
	_ = viper.BindEnv("ollama", "OLLAMA_BASE_URL")
	_ = viper.BindEnv("model", "DEFAULT_MODEL")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			Logger.Debug("config file not found", "component", "cmd.RootCmd")
		} else {
			Logger.Error("error loading config file", "error", err)
		}
	} else {
		Logger.Debug("using config file", "file", viper.ConfigFileUsed(), "component", "cmd.RootCmd")
	}

	if viper.GetBool("debug") {
		Debug = true
		initLogger()
	}
}

// initLogger initializes the logger.
func initLogger() {
	logLevel := slog.LevelWarn

	if Debug {
		logLevel = slog.LevelDebug
	}

	Logger = slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:      logLevel,
		TimeFormat: time.RFC3339,
	}))

	slog.SetDefault(Logger)
}
