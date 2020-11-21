package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const teamRepositoryTemplate = `
{{- if hasLeadingDigit .TeamName}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .TeamName}}-{{.RepoName}}
{{- end}}
# terraform import github_team_repository.{{normalizeResourceName .TeamName}}-{{.RepoName}} {{.TeamID}}:{{.RepoName}}
resource "github_team_repository" "{{normalizeResourceName .TeamName}}-{{.RepoName}}" {
  team_id    = {{.TeamID}}
  repository = "{{.RepoName}}"
  permission = "{{.Permission}}" 
}
`

func init() {
	rootCmd.AddCommand(teamRepositoryCmd)
}

var teamRepositoryCmd = &cobra.Command{
	Use:   "team-repository",
	Short: "Import organization teams repositories into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting team repository data")

		// first get teams, then for each team, get its repositories
		teams, err := getOrgTeams()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_team_repository.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, team := range teams {

			teamRepositories, err := getOrgTeamRepositorys(team)
			if err != nil {
				return
			}

			for _, repo := range teamRepositories {
				log.WithFields(logrus.Fields{
					"Team":       team.GetName(),
					"Repository": repo.GetName(),
				}).Debug("Processing team membership")

				// Figure out the team permission for this repository
				// Order is admin, maintain, push(write), triage, pull(read)
				permissions := repo.GetPermissions()
				for _, permission := range []string{"admin", "maintain", "push", "triage", "pull"} {
					if permissions[permission] {
						teamRepositoryParse(team, repo, permission, output)
						break
					}
				}

			}
		}

	},
}

func getOrgTeamRepositorys(team *github.Team) ([]*github.Repository, error) {
	opt := &github.ListOptions{PerPage: 100}

	var teamRepositories []*github.Repository
	for {
		repos, resp, err := api.Teams.ListTeamReposBySlug(ctx, orgName, team.GetSlug(), opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		teamRepositories = append(teamRepositories, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return teamRepositories, nil
}

func teamRepositoryParse(team *github.Team, repo *github.Repository, permission string, output *os.File) {
	tmpl := template.Must(template.New("team-repository").Funcs(templateFuncMap).Parse(teamRepositoryTemplate))
	err := tmpl.Execute(output,
		struct {
			Org        string
			TeamID     int64
			TeamName   string
			RepoName   string
			Permission string
		}{
			Org:        orgName,
			TeamID:     team.GetID(),
			TeamName:   team.GetName(),
			RepoName:   repo.GetName(),
			Permission: permission,
		})
	if err != nil {
		log.Error(err)
	}
}
