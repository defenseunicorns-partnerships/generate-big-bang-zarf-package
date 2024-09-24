package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/defenseunicorns/generate-big-bang-zarf-package/src/bigbang"
	"github.com/spf13/cobra"
)

var bbOpts bigbang.Opts

var rootCmd = &cobra.Command{
	Use:     "VERSION",
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a zarf.yaml and associated manifests for a Big Bang Zarf package in the current directory",
	Example: "go run main.go 2.34.0 --values-file-manifests=my-configmap.yaml,my-secret.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		bbOpts.Version = args[0]
		bbOpts.BaseDir = "."
		return bigbang.Create(cmd.Context(), bbOpts)
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

func init() {
	rootCmd.Flags().StringSliceVar(&bbOpts.ValuesFileManifests, "values-file-manifests", nil, "A comma separated list of configmap or secret manifests to pass to the Big Bang Helm Release. See https://fluxcd.io/flux/components/helm/helmreleases/#values-references")
	rootCmd.Flags().BoolVar(&bbOpts.SkipFlux, "skip-flux", false, "Skip the Flux component in the Big Bang package")
	rootCmd.Flags().BoolVar(&bbOpts.Airgap, "airgap", true, "Whether or not this package is targeting an airgap environment")
	rootCmd.Flags().StringVar(&bbOpts.Repo, "repo", "https://repo1.dso.mil/big-bang/bigbang", "The git repository to use for the Big Bang package")
}
