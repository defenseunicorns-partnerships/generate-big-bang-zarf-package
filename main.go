// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package main is the entrypoint
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/defenseunicorns/generate-big-bang-zarf-package/src/cmd"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	defer func() {
		signal.Stop(signalCh)
		cancel()
	}()
	go func() {
		select {
		case <-signalCh: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalCh // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()
	cmd.Execute(ctx)
}
