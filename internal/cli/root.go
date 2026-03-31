package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/mkh/rice-railing/internal/config"
)

var (
	cfgFile string
	verbose bool
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "rice-rail",
	Short: "Project-specific convergence toolkit",
	Long:  "rice-rail profiles repositories, generates project constitutions, builds project-specific tooling, and runs deterministic intent → tool → verify → refine cycles.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .project-toolkit/constitution.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as JSON")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildToolkitCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(baselineCmd)
	rootCmd.AddCommand(cycleCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(regenerateCmd)
	rootCmd.AddCommand(upgradeToolkitCmd)
	rootCmd.AddCommand(addSkillCmd)
	rootCmd.AddCommand(addMCPCmd)
	rootCmd.AddCommand(discoverToolsCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("constitution")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".project-toolkit")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("RICE_RAIL")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if verbose {
				fmt.Println("No constitution found. Run 'rice-rail init' first.")
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: error reading config: %v\n", err)
		}
	}
}

// paths returns resolved artifact paths respecting --config flag.
func paths() config.Paths {
	return config.ResolvePaths(cfgFile)
}
