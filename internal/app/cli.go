package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"

	"charm.land/fang/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "aiusage",
		Short: "See Claude Code and Codex usage and reset times",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 35*time.Second)
			defer cancel()
			auth, err := loadCredentials(ctx)
			if err != nil {
				return err
			}
			data, err := newClient().collect(ctx, auth, time.Now())
			if err != nil {
				return err
			}
			if jsonOutput {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(data)
			}
			_, noColor := os.LookupEnv("NO_COLOR")
			colors := term.IsTerminal(os.Stdout.Fd()) && !noColor && os.Getenv("TERM") != "dumb"
			output := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())
			_, err = fmt.Fprintln(output, render(data, colors))
			return err
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print full JSON, including null fields")
	return fang.Execute(ctx, cmd, fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM))
}
