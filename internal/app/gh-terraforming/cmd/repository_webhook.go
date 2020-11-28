package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const repositoryWebhookTemplate = `
{{- if hasLeadingDigit .RepoName}}
# WARNING this resource has an invalid identifier when used with Terraform 0.12+
# Suggestion: use this identifier instead _{{normalizeResourceName .RepoName}}-{{.ID}}
{{- end}}
# terraform import github_repository_webhook.{{normalizeResourceName .RepoName}}-{{.ID}} {{normalizeResourceName .RepoName}}/{{.ID}}
resource "github_repository_webhook" "{{normalizeResourceName .RepoName}}-{{.ID}}" {
	repository = "{{.RepoName}}"
	active     = {{.Active}}
	events     = [ {{range $i, $event := .Events}}{{if $i}}, {{end}}"{{$event}}"{{end}} ]

	configuration {
		url          = "{{.Url}}"
		content_type = "{{.ContentType}}"
		insecure_ssl = {{.InsecureSSL}}
		{{if .Secret}}secret       = "PLEASE UPDATE ME"
		{{end -}}
	}
}
`

func init() {
	rootCmd.AddCommand(repositoryWebhookCmd)
}

var repositoryWebhookCmd = &cobra.Command{
	Use:   "repository-webhook",
	Short: "Import repository webhooks into Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Getting repository webhooks data")

		// first get repositories, then for each repo, get its webhooks
		repos, err := getRepositories()
		if err != nil {
			return
		}

		output, err := os.Create(fmt.Sprintf("%s/github_repository_webhook.tf", outDirectory))
		if err != nil {
			log.Error(err)
			return
		}
		defer output.Close()

		for _, repo := range repos {

			webhooks, err := getRepositoryWebhooks(repo)
			if err != nil {
				return
			}

			for _, webhook := range webhooks {

				log.WithFields(logrus.Fields{
					"Name": *repo.Name,
				}).Debug("Processing repository")

				repositoryWebhookParse(repo, webhook, output)
			}
		}
	},
}

func getRepositoryWebhooks(repo *github.Repository) ([]*github.Hook, error) {
	opt := &github.ListOptions{PerPage: 100}

	var allWebhooks []*github.Hook
	for {
		webhooks, resp, err := api.Repositories.ListHooks(ctx, orgName, repo.GetName(), opt)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		allWebhooks = append(allWebhooks, webhooks...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Debugf("Fetching next page %d", opt.Page)
	}

	return allWebhooks, nil
}

func repositoryWebhookParse(repo *github.Repository, webhook *github.Hook, output *os.File) {
	config := webhook.Config
	if config["insecure_ssl"] == "1" {
		config["insecure_ssl"] = true
	} else {
		config["insecure_ssl"] = false
	}

	// Github will never return the actual secret
	// https://github.com/terraform-providers/terraform-provider-github/blob/6a83f820a9776793a3b3ddd6c13c176059fc983a/github/resource_github_repository_webhook.go#L115-L117
	// So let's just check if there's a secret to have a dummy value on the generated code
	// The actual secret needs to be retrived directly from the website and updated in code
	if config["secret"] != nil {
		config["secret"] = true
	} else {
		config["secret"] = false
	}

	tmpl := template.Must(template.New("repository-webhook").Funcs(templateFuncMap).Parse(repositoryWebhookTemplate))
	err := tmpl.Execute(output,
		struct {
			Org      string
			RepoName string
			ID       int64
			Active   bool
			Events   []string
			// Config
			Url         string
			ContentType string
			InsecureSSL bool
			Secret      bool
		}{
			Org:         orgName,
			RepoName:    repo.GetName(),
			ID:          webhook.GetID(),
			Active:      webhook.GetActive(),
			Events:      webhook.Events,
			Url:         config["url"].(string),
			ContentType: config["content_type"].(string),
			InsecureSSL: config["insecure_ssl"].(bool),
			Secret:      config["secret"].(bool),
		})
	if err != nil {
		log.Error(err)
	}
}
