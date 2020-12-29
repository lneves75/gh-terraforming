package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const teamTemplate = `
{{- if hasLeadingDigit .Team.Name}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .Team.Name}}
{{- end}}
# terraform import github_team.{{normalizeResourceName .Team.Name}} {{.Team.ID}}
resource "github_team" "{{normalizeResourceName .Team.Name}}" {
  name           = "{{.Team.Name}}"
  {{if .Team.Description}}description    = "{{.Team.Description}}"
  {{end -}}
  {{if .Team.Privacy}}privacy        = "{{.Team.Privacy}}"
  {{end -}}
  {{if .ParentID}}parent_team_id = "{{.ParentID}}"
  {{end -}}
  {{if .Team.LDAPDN}}ldap_dn        = "{{.Team.LDAPDN}}"
  {{end -}}
}
`

func init() {
	rootCmd.AddCommand(teamCmd)
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Import organization teams into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting team data")

		teams, err := getOrgTeams()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_team.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, team := range teams {
			log.WithFields(logrus.Fields{
				"Member": team.GetName(),
			}).Debug("Processing team")

			teamParse(team, output)
		}
	},
}

func getOrgTeams() ([]*github.Team, error) {
	opt := &github.ListOptions{PerPage: 100}

	var allTeams []*github.Team
	for {
		teams, resp, err := api.Teams.ListTeams(ctx, orgName, opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allTeams, nil
}

func teamParse(team *github.Team, output *os.File) {
	tmpl := template.Must(template.New("team").Funcs(templateFuncMap).Parse(teamTemplate))
	err := tmpl.Execute(output,
		struct {
			Org      string
			Team     github.Team
			ParentID int64
		}{
			Org:      orgName,
			Team:     *team,
			ParentID: team.GetParent().GetID(),
		})
	if err != nil {
		log.Error(err)
	}
}
