// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Defense Unicorns

// Package main is the entrypoint
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/defenseunicorns/generate-big-bang-zarf-package/src/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		first := true
		for {
			<-signalCh
			if first {
				first = false
				cancel()
				continue
			}
			os.Exit(1)
		}
	}()
	cmd.Execute(ctx)
}
