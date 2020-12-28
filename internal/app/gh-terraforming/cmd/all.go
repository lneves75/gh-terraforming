package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(allCmd)
}

var allCmd = &cobra.Command{
	Use: "all",
	Long: `Import all Github resources into Terraform.

  Currently supported resources:
  - Memberships
  - Repositories
  - Repository collaborators
  - Repository webhooks
  - Teams
  - Team memberships
  - Team repository`,

	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Importing all supported resources")

		membershipCmd.Run(cmd, args)

		repositoryCmd.Run(cmd, args)

		repositoryCollaboratorCmd.Run(cmd, args)

		repositoryWebhookCmd.Run(cmd, args)

		teamCmd.Run(cmd, args)

		teamMembershipCmd.Run(cmd, args)

		teamRepositoryCmd.Run(cmd, args)
	},
}
