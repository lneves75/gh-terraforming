# Github Terraforming
> Heavily inspired on cf-terraforming from Cloudflare

## Overview

gh-terraforming is a command line utility to facilitate terraforming your existing Github resources. It does this by using a personal access token to make requests to the [Github API](https://docs.github.com/en/free-pro-team@latest/rest) and then generate Terraform configuration that can be used with the [Terraform Github provider](https://registry.terraform.io/providers/hashicorp/github/latest).

This tool is ideal if you already have Github resources defined but want to start managing them via Terraform, and don't want to spend the time to manually write the Terraform configuration to describe them.

## Usage

```
Usage:
  gh-terraforming [command]

Available Commands:
  repository     Import Repository data into Terraform
  version                Print the version number of gh-terraforming

Flags:
  -o, --organization string   Use specific organization for import
  -t, --token string          Token generated on the 'Personal access tokens' page, under 'Developer settings'. See: https://github.com/settings/tokens
  -d, --out-dir string        Location where the resource files will be written to (defaults to PWD)
  -h, --help                  help for gh-terraforming
  -l, --loglevel string       Specify logging level: (trace, debug, info, warn, error, fatal, panic)
  -v, --verbose               Specify verbose output (same as setting log level to debug)

Use "gh-terraforming [command] --help" for more information about a command.
```

## Authentication

As Github will deprecate basic auth we use personal access tokens for authentication

It can be generated on the [personal access token page](https://github.com/settings/tokens).

**A note on storing your credentials securely:** It's a good practice to store your Github token as an environment variable as demonstrated below.

```bash
export GITHUB_TOKEN='<token value>'

# now call gh-terraforming, e.g.
gh-terraforming --organization acme repository
```

gh-terraforming supports the following environment variables:
* GITHUB_TOKEN - Token based authentication
* GITHUB_ORGANIZATION - Organization to use in the api requests

## Example usage

```gh-terraforming --organization acme repository```

will make requests to the Github API and result in a valid Terraform configuration representing the **resource** you requested.
The required terraform import command is also generated and placed as a comment immediately before the generated resource code:

```
# terraform import github_repository.gh-terraforming gh-terraforming
resource "github_repository" "gh-terraforming" {
  name = "gh-terraforming"
  description = "Command line utility to facilitate terraforming existing Github resources"
  visibility = "public"
  has_downloads = true
  has_issues = true
  has_projects = true
  has_wiki = true
}
```

## Controlling output and verbose mode
By default, gh-terraforming will not output any log type messages to stdout when run, so as to not pollute your generated Terraform config files and to allow you to cleanly redirect gh-terraforming output to existing Terraform configs.

However, it can be useful when debugging issues to specify a logging level, like so:

```
gh-terraforming --organization acme -l debug repository

DEBU[0000] Initializing go-github                        Organization=acmd Token="*************1234"
DEBU[0000] Getting repository data
DEBU[0001] Fetching next page 2
DEBU[0002] Processing repository                         Name=gh-terraforming
```

For convenience, you can set the verbose flag, which is functionally equivalent to setting a log level of debug:

```
gh-terraforming --organization acme -v repository
```

## Prerequisites
* A Github organization you have access to
* A valid Github personal access token and sufficient permissions to access the resources you are requesting via the API
* A working [installation of Go](https://golang.org/doc/install) at least v1.12.x.

## Installation

```bash
$ go get -u github.com/lneves75/gh-terraforming/...
```
This will fetch the gh-terraforming tool as well as its dependencies, updating them as necessary, build and install the package in your `$GOPATH` (usually `~/go/bin`). You can check your current GOPATH by running:

```bash
$ go env | grep GOPATH
```

## Supported resources

The following resources can be downloaded into [Terraform HCL format](https://www.terraform.io/docs/configuration/syntax.html) right now. Support across the remaining commands will be added over time.

| Resource | Generating HCL |
|----------|----------------|
| [repository](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository) | ✔️ |
| [actions_secret](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/actions_secret) | ✖️ |
| [branch](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/branch) | ✖️ |
| [branch_protection](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/branch_protection) | ✖️ |
| [issue_label](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/issue_label) | ✖️ |
| [membership](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/membership) | ✖️ |
| [organization_block](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/organization_block) | ✖️ |
| [organization_project](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/organization_project) | ✖️ |
| [organization_webhook](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/organization_webhook) | ✖️ |
| [project_column](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/project_column) | ✖️ |
| [repository_collaborator](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository_collaborator) | ✖️ |
| [repository_deploy_key](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository_deploy_key) | ✖️ |
| [repository_file](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository_file) | ✖️ |
| [repository_project](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository_project) | ✖️ |
| [repository_webhook](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/repository_webhook) | ✖️ |
| [team](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/team) | ✖️ |
| [team_membership](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/team_membership) | ✖️ |
| [team_repository](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/team_repository) | ✖️ |
| [team_sync_group_mapping](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/team_sync_group_mapping) | ✖️ |
| [user_gpg_key](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/user_gpg_key) | ✖️ |
| [user_invitation_accepter](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/user_invitation_accepter) | ✖️ |
| [user_ssh_key](https://registry.terraform.io/providers/hashicorp/github/latest/docs/resources/user_ssh_key) | ✖️ |