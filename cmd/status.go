package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := edwardClient.Status(args, *statusFlags.all)
		return errors.WithStack(err)
	},
}

var statusFlags struct {
	all *bool
}

func init() {
	RootCmd.AddCommand(statusCmd)

	statusFlags.all = statusCmd.Flags().BoolP("all", "a", false, "Show status for all services, even those in other Edward configs.")
}
