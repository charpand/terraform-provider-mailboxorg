# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Provider configuration through `username`, `password`, `base_url` and `timeout`, or the matching `MAILBOXORG_*` environment variables
- `mailboxorg_domain` data source backed by `domain.get`
- JSON-RPC client with a lazily created, shared session that is renewed once when the API reports it invalid, for read calls only
