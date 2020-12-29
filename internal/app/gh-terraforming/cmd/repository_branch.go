package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const repositoryBranchTemplate = `
{{- if hasLeadingDigit .Repo}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .Repo}}-{{.Branch}}
{{- end}}
# terraform import github_repository_branch.{{normalizeResourceName .Repo}}-{{.Branch}} {{normalizeResourceName .Repo}}:{{.Branch}}
resource "github_repository_branch" "{{normalizeResourceName .Repo}}-{{.Branch}}" {
	repository = "{{.Repo}}"
	branch     = "{{.Branch}}"
}
`

func init() {
	rootCmd.AddCommand(repositoryBranchCmd)
}

var repositoryBranchCmd = &cobra.Command{
	Use:   "repository-branch",
	Short: "Import repository branches into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting repository branches data")

		// first get repositories, then for each repo, get its branches
		repos, err := getRepositories()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_repository_branch.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		//TODO: remove this warning once we support getting source branch and sha
		fmt.Fprintln(output, "# WARNING: for now the tool will not output source_branch and source_sha therefore assuming their default values")
		fmt.Fprintln(output, "# as per provider documentation: https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/branch")
		fmt.Fprintln(output, "# Support for getting the actual values will be added eventually")

		for _, repo := range repos {

			branches, err := getRepositoryBranches(repo)
			if err != nil {
				return
			}

			for _, branch := range branches {

				log.WithFields(logrus.Fields{
					"Repository": repo.GetName(),
					"Branch":     branch.GetName(),
				}).Debug("Processing repository")

				repositoryBranchParse(repo, branch, output)
			}
		}
	},
}

func getRepositoryBranches(repo *github.Repository) ([]*github.Branch, error) {
	opt := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := api.Repositories.ListBranches(ctx, orgName, repo.GetName(), opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allBranches, nil
}

func repositoryBranchParse(repo *github.Repository, branch *github.Branch, output *os.File) {

	tmpl := template.Must(template.New("repository-branch").Funcs(templateFuncMap).Parse(repositoryBranchTemplate))
	err := tmpl.Execute(output,
		struct {
			Org    string
			Repo   string
			Branch string
		}{
			Org:    orgName,
			Repo:   repo.GetName(),
			Branch: branch.GetName(),
		})
	if err != nil {
		log.Error(err)
	}
}
