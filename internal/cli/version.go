package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print rice-rail version",
	Run: func(cmd *cobra.Command, args []string) {
		version := "dev"
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
		}
		fmt.Printf("rice-rail %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
