package cli

import "github.com/spf13/cobra"

// coursesCmd is an alias for lessons (spec uses "courses <path-id>").
func (a *App) coursesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "courses <path-slug>",
		Short: "List all lessons in a learning path (alias for lessons)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lessons, err := a.client.Lessons(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			if a.limit > 0 && len(lessons) > a.limit {
				lessons = lessons[:a.limit]
			}
			return a.renderOrEmpty(lessons, len(lessons))
		},
	}
}

func (a *App) infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Print site stats (path count)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := a.client.Info(cmd.Context())
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(info)
		},
	}
}
