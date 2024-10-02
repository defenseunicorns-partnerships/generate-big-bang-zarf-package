// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package helm contains operations for working with helm charts.
package helm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/defenseunicorns/generate-big-bang-zarf-package/internal/git"

	"github.com/defenseunicorns/pkg/helpers/v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/client-go/util/homedir"

	"github.com/zarf-dev/zarf/src/config"
	"github.com/zarf-dev/zarf/src/config/lang"
	"github.com/zarf-dev/zarf/src/pkg/message"
	"github.com/zarf-dev/zarf/src/pkg/utils"
)

// PackageChartFromLocalFiles creates a chart archive from a path to a chart on the host os.
func (h *Helm) PackageChartFromLocalFiles(ctx context.Context, cosignKeyPath string) error {
	spinner := message.NewProgressSpinner("Processing helm chart %s:%s from %s", h.chart.Name, h.chart.Version, h.chart.LocalPath)
	defer spinner.Stop()

	// Load and validate the chart
	cl, _, err := h.loadAndValidateChart(h.chart.LocalPath)
	if err != nil {
		return err
	}

	// Handle the chart directory or tarball
	var saved string
	temp := filepath.Join(h.chartPath, "temp")
	if _, ok := cl.(loader.DirLoader); ok {
		err = h.buildChartDependencies()
		if err != nil {
			return fmt.Errorf("unable to build dependencies for the chart: %w", err)
		}

		client := action.NewPackage()

		client.Destination = temp
		saved, err = client.Run(h.chart.LocalPath, nil)
	} else {
		saved = filepath.Join(temp, filepath.Base(h.chart.LocalPath))
		err = helpers.CreatePathAndCopy(h.chart.LocalPath, saved)
	}
	defer os.RemoveAll(temp)

	if err != nil {
		return fmt.Errorf("unable to save the archive and create the package %s: %w", saved, err)
	}

	// Finalize the chart
	err = h.finalizeChartPackage(ctx, saved, cosignKeyPath)
	if err != nil {
		return err
	}

	spinner.Success()

	return nil
}

// PackageChartFromGit is a special implementation of chart archiving that supports the https://p1.dso.mil/#/products/big-bang/ model.
// needed
func (h *Helm) PackageChartFromGit(ctx context.Context, cosignKeyPath string) error {
	spinner := message.NewProgressSpinner("Processing helm chart %s", h.chart.Name)
	defer spinner.Stop()

	// Retrieve the repo containing the chart
	gitPath, err := DownloadChartFromGitToTemp(ctx, h.chart.URL)
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitPath)

	// Set the directory for the chart and package it
	h.chart.LocalPath = filepath.Join(gitPath, h.chart.GitPath)
	return h.PackageChartFromLocalFiles(ctx, cosignKeyPath)
}

// DownloadChartFromGitToTemp downloads a chart from git into a temp directory
func DownloadChartFromGitToTemp(ctx context.Context, url string) (string, error) {
	path, err := utils.MakeTempDir(config.CommonOptions.TempDirectory)
	if err != nil {
		return "", fmt.Errorf("unable to create tmpdir: %w", err)
	}
	repository, err := git.Clone(ctx, path, url, true)
	if err != nil {
		return "", err
	}
	return repository.Path(), nil
}

func (h *Helm) finalizeChartPackage(ctx context.Context, saved, cosignKeyPath string) error {
	// Ensure the name is consistent for deployments
	destinationTarball := StandardName(h.chartPath, h.chart) + ".tgz"
	err := os.Rename(saved, destinationTarball)
	if err != nil {
		return fmt.Errorf("unable to save the final chart tarball: %w", err)
	}

	err = h.packageValues(ctx, cosignKeyPath)
	if err != nil {
		return fmt.Errorf("unable to process the values for the package: %w", err)
	}
	return nil
}

func (h *Helm) packageValues(ctx context.Context, cosignKeyPath string) error {
	for valuesIdx, path := range h.chart.ValuesFiles {
		dst := StandardValuesName(h.valuesPath, h.chart, valuesIdx)

		if helpers.IsURL(path) {
			if err := utils.DownloadToFile(ctx, path, dst, cosignKeyPath); err != nil {
				return fmt.Errorf(lang.ErrDownloading, path, err.Error())
			}
		} else {
			if err := helpers.CreatePathAndCopy(path, dst); err != nil {
				return fmt.Errorf("unable to copy chart values file %s: %w", path, err)
			}
		}
	}

	return nil
}

// buildChartDependencies builds the helm chart dependencies
func (h *Helm) buildChartDependencies() error {
	// Download and build the specified dependencies
	regClient, err := registry.NewClient(registry.ClientOptEnableCache(true))
	if err != nil {
		return fmt.Errorf("unable to create a new registry client: %w", err)
	}

	h.settings = cli.New()
	defaultKeyring := filepath.Join(homedir.HomeDir(), ".gnupg", "pubring.gpg")
	if v, ok := os.LookupEnv("GNUPGHOME"); ok {
		defaultKeyring = filepath.Join(v, "pubring.gpg")
	}

	man := &downloader.Manager{
		Out:            &message.DebugWriter{},
		ChartPath:      h.chart.LocalPath,
		Getters:        getter.All(h.settings),
		RegistryClient: regClient,

		RepositoryConfig: h.settings.RepositoryConfig,
		RepositoryCache:  h.settings.RepositoryCache,
		Debug:            false,
		Verify:           downloader.VerifyIfPossible,
		Keyring:          defaultKeyring,
	}

	// Build the deps from the helm chart
	err = man.Build()
	var notFoundErr *downloader.ErrRepoNotFound
	if errors.As(err, &notFoundErr) {
		// If we encounter a repo not found error point the user to `zarf tools helm repo add`
		message.Warnf("%s. Please add the missing repo(s) via the following:", notFoundErr.Error())
		for _, repository := range notFoundErr.Repos {
			message.ZarfCommand(fmt.Sprintf("tools helm repo add <your-repo-name> %s", repository))
		}
		return err
	}
	if err != nil {
		message.ZarfCommand("tools helm dependency build --verify")
		message.Warnf("Unable to perform a rebuild of Helm dependencies: %s", err.Error())
		return err
	}
	return nil
}

func (h *Helm) loadAndValidateChart(location string) (loader.ChartLoader, *chart.Chart, error) {
	// Validate the chart
	cl, err := loader.Loader(location)
	if err != nil {
		return cl, nil, fmt.Errorf("unable to load the chart from %s: %w", location, err)
	}

	chart, err := cl.Load()
	if err != nil {
		return cl, chart, fmt.Errorf("validation failed for chart from %s: %w", location, err)
	}

	return cl, chart, nil
}