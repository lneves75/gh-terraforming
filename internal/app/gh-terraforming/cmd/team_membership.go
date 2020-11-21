package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const teamMembershipTemplate = `
{{- if hasLeadingDigit .TeamName}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .TeamName}}-{{.UserName}}
{{- end}}
# terraform import github_team_membership.{{normalizeResourceName .TeamName}}-{{.UserName}} {{.TeamID}}:{{.UserName}}
resource "github_team_membership" "{{normalizeResourceName .TeamName}}-{{.UserName}}" {
  team_id = {{.TeamID}}
  username = "{{.UserName}}"
  {{if .Role}}role = "{{.Role}}" 
  {{end -}}
}
`

func init() {
	rootCmd.AddCommand(teamMembershipCmd)
}

var teamMembershipCmd = &cobra.Command{
	Use:   "team-membership",
	Short: "Import organization teams memberships into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting team membership data")

		// first get teams, then for each team, get its members
		teams, err := getOrgTeams()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_team_membership.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, role := range []string{"maintainer", "member"} {

			for _, team := range teams {

				teamMembers, err := getOrgTeamMemberships(team, role)
				if err != nil {
					return
				}

				for _, teamMember := range teamMembers {
					log.WithFields(logrus.Fields{
						"Team":   teamMember.GetName(),
						"Member": teamMember.GetName(),
					}).Debug("Processing team membership")

					teamMembershipParse(team, teamMember, role, output)
				}
			}
		}

	},
}

func getOrgTeamMemberships(team *github.Team, role string) ([]*github.User, error) {
	opt := &github.TeamListTeamMembersOptions{
		Role:        role,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var teamMembers []*github.User
	for {
		users, resp, err := api.Teams.ListTeamMembersBySlug(ctx, orgName, team.GetSlug(), opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		teamMembers = append(teamMembers, users...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return teamMembers, nil
}

func teamMembershipParse(team *github.Team, user *github.User, role string, output *os.File) {
	tmpl := template.Must(template.New("team-membership").Funcs(templateFuncMap).Parse(teamMembershipTemplate))
	err := tmpl.Execute(output,
		struct {
			Org      string
			TeamID   int64
			TeamName string
			UserName string
			Role     string
		}{
			Org:      orgName,
			TeamID:   team.GetID(),
			TeamName: team.GetName(),
			UserName: user.GetLogin(),
			Role:     role,
		})
	if err != nil {
		log.Error(err)
	}
}
