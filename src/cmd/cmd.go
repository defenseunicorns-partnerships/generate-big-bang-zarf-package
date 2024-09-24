package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/defenseunicorns/generate-big-bang-zarf-package/src/bigbang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "zarf COMMAND",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute(ctx context.Context) {
	_, err := rootCmd.ExecuteContextC(ctx)
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}

var bbOpts bigbang.Opts

var bigBangGenerateCommand = &cobra.Command{
	Use:     "big-bang VERSION",
	Aliases: []string{"bb"},
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a zarf.yaml and associated manifests for a Big Bang Zarf package in the current directory",
	Example: "zarf dev generate big-bang 2.3.4 --values-file-manifests=my-configmap.yaml,my-secret.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		bbOpts.Version = args[0]
		bbOpts.BaseDir = "."
		return bigbang.Create(cmd.Context(), bbOpts)
	},
}
