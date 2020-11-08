package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const membershipTemplate = `
{{- if hasLeadingDigit .Username}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .Username}}
{{- end}}
# terraform import github_membership.{{normalizeResourceName .Username}} {{.Org}}:{{normalizeResourceName .Username}}
resource "github_membership" "{{normalizeResourceName .Username}}" {
  username = "{{.Username}}"
  role     = "{{.Role}}"
}
`

func init() {
	rootCmd.AddCommand(membershipCmd)
}

var membershipCmd = &cobra.Command{
	Use:   "membership",
	Short: "Import organization members into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting membership data")

		members, err := getOrgMembers()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_membership.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, member := range members {
			log.WithFields(logrus.Fields{
				"Member": member.GetLogin(),
			}).Debug("Processing membership")

			membershipParse(member, output)
		}
	},
}

func getOrgMembers() ([]*github.User, error) {
	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allMembers []*github.User
	for {
		members, resp, err := api.Organizations.ListMembers(ctx, orgName, opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allMembers, nil
}

func membershipParse(user *github.User, output *os.File) {
	// Get the organization role for this member
	membership, _, err := api.Organizations.GetOrgMembership(ctx, user.GetLogin(), orgName)
	if err != nil {
		log.Error(err)
		return
	}

	tmpl := template.Must(template.New("membership").Funcs(templateFuncMap).Parse(membershipTemplate))
	err = tmpl.Execute(output,
		struct {
			Org      string
			Username string
			Role     string
		}{
			Org:      orgName,
			Username: user.GetLogin(),
			Role:     membership.GetRole(),
		})
	if err != nil {
		log.Error(err)
	}
}
