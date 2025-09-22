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
	Model      string
	Logger     *slog.Logger
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "ghost",
	Short: "A cyberpunk inspired AI assistant.",
	Long:  "Ghost is a CLI tool that provides AI-powered assistance directly in your terminal inspired by cyberpunk media.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		Logger.Error("error running ghost command", slog.String("component", "rootCmd"))
		os.Exit(1)
	}
}

// init initializes the application.
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.ghost.yaml)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVar(&Model, "model", "", "LLM model to use")

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	initLogger()

	if err := godotenv.Load(); err != nil {
		Logger.Debug(".env file not found, using environment variables", "component", "rootCmd")
	} else {
		Logger.Debug(".env file loaded successfully", "component", "rootCmd")
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
		viper.SetConfigType(".yaml")
	}

	viper.AutomaticEnv()
	viper.BindEnv("ollama_base_url", "OLLAMA_BASE_URL")
	viper.BindEnv("model", "DEFAULT_MODEL")

	if err := viper.ReadInConfig(); err == nil {
		Logger.Debug("using config file", "file", viper.ConfigFileUsed(), "component", "rootCmd")
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

	Logger = slog.New(slogcolor.NewHandler(askCmd.ErrOrStderr(), &slogcolor.Options{
		Level:      logLevel,
		TimeFormat: time.RFC3339,
	}))

	slog.SetDefault(Logger)
}
