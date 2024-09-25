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

// Execute runs the root command, it's called from main.go
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
	// In Zarf the default kube version using ld flags. For simplicity we aren't introducing ld flags in this project
	// https://github.com/zarf-dev/zarf/blob/dd40adbd54412dd096ab88b5ce48cdeaa261e2fe/.goreleaser.yaml#L24-L25
	// Instead Kube version is set very high by default so we always met the chart minimum version requirements
	// There is an option to override in case big bang ever adds a maximum version requirement
	rootCmd.Flags().StringVar(&bbOpts.KubeVersion, "kube-version", "1.99.0", "Override the default KubeVersion used during the helm template portion of generate")
}
