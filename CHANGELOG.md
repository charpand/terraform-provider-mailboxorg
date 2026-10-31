# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- CI jobs run on CRACI runners (`runs-on: craci`, `cracicorp/setup` with the x64 `default` image at size 2) instead of GitHub's `ubuntu-26.04-arm` runners
- CI installs Go, golangci-lint, Terraform and Node through `jdx/mise-action` at the versions in `mise.toml` and `mise.lock`, so CI uses the same tools as a local checkout; before, the lint job had no golangci-lint and `./scripts/lint` fell back to `go vet`
- CI checks that `go.mod` and `go.sum` are tidy (`go mod tidy -diff`) and that `./scripts/docs` leaves no new or changed files in `docs/` or `templates/`; the test job no longer runs `go mod tidy`
- CI uses one concurrency group per ref: a newer push cancels a pull request's older run, while runs on `main` and on tags always complete; checkouts do not keep the token, and GoReleaser stays on v2

### Removed
- The `snyk` CI job; Snyk scans pull requests and `main` through the Snyk GitHub App (`security/snyk` and `code/snyk` statuses), so `SNYK_TOKEN` is no longer used

## [0.2.0] - 2026-10-17

### Added
- `mailboxorg_domains` data source backed by `domain.list`
- `mailboxorg_mail` data source backed by `mail.get` (password and postal address data are not exposed)
- `mailboxorg_mails` data source backed by `mail.list`, reading every page
- Documented that `mailboxorg_mails` `query` is scoped to the login (empty for a reseller login)
- API error messages now include the API's detail text when it sends one

## [0.1.0] - 2026-10-17

### Added
- Provider configuration through `username`, `password`, `base_url` and `timeout`, or the matching `MAILBOXORG_*` environment variables
- `mailboxorg_domain` data source backed by `domain.get`
- JSON-RPC client with a lazily created, shared session that is renewed once when the API reports it invalid, for read calls only

### Changed
- `mailboxorg_domain` reports API error 8 as "Domain not found or not manageable", since the API does not distinguish a missing domain from one the account may not manage
- Documented that `mailboxorg_domain` requires an API user that manages the domain
- Documented using several logins in one state with provider aliases
