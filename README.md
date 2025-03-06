![Baton Logo](./baton-logo.png)

# `baton-panda-doc` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-panda-doc.svg)](https://pkg.go.dev/github.com/conductorone/baton-panda-doc) ![main ci](https://github.com/conductorone/baton-panda-doc/actions/workflows/main.yaml/badge.svg)

`baton-panda-doc` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Getting Started

## Prerequisites

To obtain the necessary API key, in your PandaDoc account, go to Dev Center, Configuration, under API keys you will be able to generate production or Sandbox key. For more information visit: [API-Key Documentation](https://developers.pandadoc.com/reference/api-key-authentication-process)

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-panda-doc
baton-panda-doc
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-panda-doc:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-panda-doc/cmd/baton-panda-doc@main

baton-panda-doc

baton resources
```

# Data Model

`baton-panda-doc` will pull down information about the following resources:
- Users
- Workspaces
- Roles

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-panda-doc` Command Line Usage

```
baton-panda-doc

Usage:
  baton-panda-doc [flags]
  baton-panda-doc [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-key string               required: The API key for your PandaDoc account ($BATON_API_KEY)
      --domain string                Optional: Set to 'eu' for Europe API instance ($BATON_API_DOMAIN)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-panda-doc
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-panda-doc

Use "baton-panda-doc [command] --help" for more information about a command.
```
