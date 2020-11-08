package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const repositoryTemplate = `
{{- if hasLeadingDigit .Repository.Name}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .Repository.Name}}
{{- end}}
# terraform import github_repository.{{normalizeResourceName .Repository.Name}} {{normalizeResourceName .Repository.Name}}
resource "github_repository" "{{normalizeResourceName .Repository.Name}}" {
  name = "{{.Repository.Name}}"
  {{if .Repository.Description}}description = "{{.Repository.Description}}" 
  {{end -}}
  {{if .Repository.Homepage}}homepage_url = "{{.Repository.Homepage}}" 
  {{end -}}
  {{ if not .Repository.Visibility }}private = {{.Repository.Private}} {{else -}}visibility = {{.Repository.Visibility}} 
  {{end}}
  {{if .Repository.HasDownloads }}has_downloads = {{.Repository.HasDownloads}} 
  {{end -}}
  {{ if .Repository.HasIssues }}has_issues = {{.Repository.HasIssues}} 
  {{end -}}
  {{ if .Repository.HasProjects }}has_projects = {{.Repository.HasProjects}} 
  {{end -}}
  {{ if .Repository.HasWiki }}has_wiki = {{.Repository.HasWiki}} 
  {{end -}}
  {{ if .Repository.IsTemplate }}is_template = {{.Repository.IsTemplate}} 
  {{end -}}
  {{ if .Repository.AllowMergeCommit }}allow_merge_commit = {{.Repository.AllowMergeCommit}} 
  {{end -}}
  {{ if .Repository.AllowSquashMerge }}allow_squash_merge = {{.Repository.AllowSquashMerge}} 
  {{end -}}
  {{ if .Repository.AllowRebaseMerge }}allow_rebase_merge = {{.Repository.AllowRebaseMerge}} 
  {{end -}}
  {{ if .Repository.DeleteBranchOnMerge }}delete_branch_on_merge = {{.Repository.DeleteBranchOnMerge}} 
  {{end -}}
  {{ if .Repository.AutoInit }}auto_init = {{.Repository.AutoInit}} 
  {{end -}}
  {{ if .Repository.LicenseTemplate }}license_template = "{{.Repository.LicenseTemplate}}" 
  {{end -}}
  {{ if .Repository.GitignoreTemplate }}gitignore_template = "{{.Repository.GitignoreTemplate}}" 
  {{end -}}
  {{ if not .Repository.Archived }}archived = {{.Repository.Archived}} 
  {{end -}}
  {{ if .Repository.Topics }}topics = [ {{range .Repository.Topics}}"{{.}}",{{end}} ]
  {{end -}}
}
`

func init() {
	rootCmd.AddCommand(repositoryCmd)
}

var repositoryCmd = &cobra.Command{
	Use:   "repository",
	Short: "Import repository resources into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting repository data")

		repos, err := getRepositories()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_repository.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, repo := range repos {
			log.WithFields(logrus.Fields{
				"Name": *repo.Name,
			}).Debug("Processing repository")

			repositoryParse(repo, output)
		}
	},
}

func getRepositories() ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := api.Repositories.ListByOrg(ctx, orgName, opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allRepos, nil
}

func repositoryParse(repo *github.Repository, output *os.File) {
	tmpl := template.Must(template.New("repository").Funcs(templateFuncMap).Parse(repositoryTemplate))
	err := tmpl.Execute(output,
		struct {
			Org        string
			Repository github.Repository
		}{
			Org:        orgName,
			Repository: *repo,
		})
	if err != nil {
		log.Error(err)
	}
}
