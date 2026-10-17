# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `mailboxorg_domains` data source backed by `domain.list`
- `mailboxorg_mail` data source backed by `mail.get` (password and postal address data are not exposed)
- Provider configuration through `username`, `password`, `base_url` and `timeout`, or the matching `MAILBOXORG_*` environment variables
- `mailboxorg_domain` data source backed by `domain.get`
- JSON-RPC client with a lazily created, shared session that is renewed once when the API reports it invalid, for read calls only

### Changed
- `mailboxorg_domain` reports API error 8 as "Domain not found or not manageable", since the API does not distinguish a missing domain from one the account may not manage
- Documented that `mailboxorg_domain` requires an API user that manages the domain
- Documented using several logins in one state with provider aliases
