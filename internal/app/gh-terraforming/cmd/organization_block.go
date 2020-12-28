package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const organizationBlockTemplate = `
{{- if hasLeadingDigit .Username}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .Username}}
{{- end}}
# terraform import github_organization_block.{{normalizeResourceName .Username}} {{normalizeResourceName .Username}}
resource "github_organization_block" "{{normalizeResourceName .Username}}" {
  username = "{{.Username}}"
}
`

func init() {
	rootCmd.AddCommand(organizationBlockCmd)
}

var organizationBlockCmd = &cobra.Command{
	Use:   "organization-block",
	Short: "Import organization blocked users into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting organization blocked users data")

		users, err := getorganizationBlockedUsers()
		if err != nil {
			return
		}

		if len(users) == 0 {
			log.Info("Nothing found")
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_organization_blocks.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, user := range users {

			log.WithFields(logrus.Fields{
				"User": user.GetLogin(),
			}).Debug("Processing user block")

			organizationBlockParse(user, output)
		}
	},
}

func getorganizationBlockedUsers() ([]*github.User, error) {
	opt := &github.ListOptions{PerPage: 100}

	var allUsers []*github.User
	for {
		users, resp, err := api.Organizations.ListBlockedUsers(ctx, orgName, opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allUsers, nil
}

func organizationBlockParse(user *github.User, output *os.File) {

	tmpl := template.Must(template.New("organization-block").Funcs(templateFuncMap).Parse(organizationBlockTemplate))
	err := tmpl.Execute(output,
		struct {
			Org      string
			Username string
		}{
			Org:      orgName,
			Username: user.GetLogin(),
		})
	if err != nil {
		log.Error(err)
	}
}
