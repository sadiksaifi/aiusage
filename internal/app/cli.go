package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"syscall"
	"time"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

func Run(ctx context.Context) error {
	var jsonOutput, upgradeRequested bool
	cmd := &cobra.Command{
		Use:   "aiusage",
		Short: "See Claude Code and Codex usage and reset times",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if upgradeRequested {
				if cmd.Flags().NFlag() > 1 {
					return fmt.Errorf("use --upgrade on its own, without other flags")
				}
				return upgrade(cmd.Context(), cmd.OutOrStdout())
			}
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
			width, dark := 120, true
			if term.IsTerminal(os.Stdout.Fd()) {
				if columns, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
					width = columns
				}
				if colors {
					dark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
				}
			}
			output := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())
			_, err = fmt.Fprintln(output, render(data, colors, width, dark))
			return err
		},
	}
	cmd.Flags().BoolVar(&upgradeRequested, "upgrade", false, "Upgrade to the latest stable release with SHA-256 verification")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print full JSON, including null fields")
	return fang.Execute(ctx, cmd, fang.WithVersion(currentVersion()), fang.WithNotifySignal(os.Interrupt, syscall.SIGTERM))
}
