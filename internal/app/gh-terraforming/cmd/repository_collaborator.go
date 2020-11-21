package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const repositoryCollaboratorTemplate = `
{{- if hasLeadingDigit .RepoName}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .RepoName}}-{{.UserName}}
{{- end}}
# terraform import github_repository_collaborator.{{normalizeResourceName .RepoName}}-{{.UserName}} {{.RepoName}}:{{.UserName}}
resource "github_repository_collaborator" "{{normalizeResourceName .RepoName}}-{{.UserName}}" {
  repository = "{{.RepoName}}"
  username   = "{{.UserName}}"
  permission = "{{.Permission}}" 
}
`

func init() {
	rootCmd.AddCommand(repositoryCollaboratorCmd)
}

var repositoryCollaboratorCmd = &cobra.Command{
	Use:   "repository-collaborator",
	Short: "Import organization repository collaborators into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting repository collaborator data")

		// first get repositories, then for each repo, get its collaborators
		repos, err := getRepositories()
		if err != nil {
			return
		}

		// generate the code to two separate files, one for external collaborators and another for org members
		output := make(map[string]*os.File)

		// file for external collaborators
		output["outside"], err = os.Create(fmt.Sprintf("%s/github_repository_external_collaborator.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output["outside"].Close()

		// file for organization members
		output["direct"], err = os.Create(fmt.Sprintf("%s/github_repository_collaborator.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output["direct"].Close()

		var externalCollaborators []*github.User
		for _, repo := range repos {

			for _, affiliation := range []string{"outside", "direct"} {

				collaborators, err := getOrgRepositoryCollaborators(repo, affiliation)
				if err != nil {
					return
				}

				/*
					github's api will return all the repositories collaborators when we query
					with direct affiliation but we want to exclude the outside ones from that list,
					so let's save the outside collaborators in a separate variable
				*/
				if affiliation == "outside" {
					externalCollaborators = collaborators

				} else {

					// Now we iterate the list of outside collaborators and exclude them from the direct list
					for _, external := range externalCollaborators {

						var updatedCollaborators []*github.User
						for _, collaborator := range collaborators {

							if external.GetLogin() != collaborator.GetLogin() {
								updatedCollaborators = append(updatedCollaborators, collaborator)
							}
						}

						collaborators = updatedCollaborators
					}

				}

				for _, collaborator := range collaborators {

					log.WithFields(logrus.Fields{
						"Repository":   repo.GetName(),
						"Collaborator": collaborator.GetLogin(),
						"Affiliation":  affiliation,
					}).Debug("Processing repository collaborator")

					// Figure out the collaborator permission for this repository
					// Order is admin, maintain, push(write), triage, pull(read)
					permissions := collaborator.GetPermissions()
					for _, permission := range []string{"admin", "maintain", "push", "triage", "pull"} {
						if permissions[permission] {
							repositoryCollaboratorParse(repo, collaborator, permission, output[affiliation])
							break
						}
					}
				}
			}
		}
	},
}

func getOrgRepositoryCollaborators(repo *github.Repository, affiliation string) ([]*github.User, error) {
	opt := &github.ListCollaboratorsOptions{
		Affiliation: affiliation,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var repoCollaborators []*github.User
	for {
		repos, resp, err := api.Repositories.ListCollaborators(ctx, orgName, repo.GetName(), opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		repoCollaborators = append(repoCollaborators, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return repoCollaborators, nil
}

func repositoryCollaboratorParse(repo *github.Repository, collaborator *github.User, permission string, output *os.File) {
	tmpl := template.Must(template.New("repository-collaborator").Funcs(templateFuncMap).Parse(repositoryCollaboratorTemplate))
	err := tmpl.Execute(output,
		struct {
			Org        string
			RepoName   string
			UserName   string
			Permission string
		}{
			Org:        orgName,
			RepoName:   repo.GetName(),
			UserName:   collaborator.GetLogin(),
			Permission: permission,
		})
	if err != nil {
		log.Error(err)
	}
}
