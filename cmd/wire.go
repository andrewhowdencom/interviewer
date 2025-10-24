//go:build wireinject
// +build wireinject

package cmd

import (
	"github.com/andrewhowdencom/vox/internal/ports/cli"
	"github.com/andrewhowdencom/vox/internal/ports/web"
	"github.com/google/wire"
	"github.com/spf13/cobra"
)

func initializeApp() (*cobra.Command, error) {
	wire.Build(
		newRootCmd,
		newInterviewCmd,
		provideStartCmd,
		web.NewServeCmd,
		cli.NewDebugCmd,
	)
	return &cobra.Command{}, nil
}
