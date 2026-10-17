# Terraform provider for mailbox.org

A Terraform provider for [mailbox.org](https://mailbox.org), built on the
[JSON-RPC API](https://api.mailbox.org/v1/doc/methods/index.html).

**Status:** pre-release. The first release (0.1.0) is read-only.

## Supported

| Type | Name | API method |
|---|---|---|
| Data source | `mailboxorg_domain` | `domain.get` |

## Usage

```hcl
terraform {
  required_providers {
    mailboxorg = {
      source = "charpand/mailboxorg"
    }
  }
}

provider "mailboxorg" {
  # or MAILBOXORG_USERNAME / MAILBOXORG_PASSWORD
  username = var.mailboxorg_username
  password = var.mailboxorg_password
}

data "mailboxorg_domain" "example" {
  domain = "example.com"
}
```

| Attribute | Environment variable | Default |
|---|---|---|
| `username` | `MAILBOXORG_USERNAME` | |
| `password` | `MAILBOXORG_PASSWORD` | |
| `base_url` | `MAILBOXORG_BASE_URL` | `https://api.mailbox.org/v1/` |
| `timeout` (seconds) | `MAILBOXORG_TIMEOUT` | `60` |

The API user decides what can be read: `mailboxorg_domain` only works for
domains the user manages (verified with a login that manages the domain). A
reseller manages accounts, and domains belong to an account, so a login that
does not manage the domain gets API error 8 (see below). Other API errors are
reported unchanged.

## Known limitations

- The API has no "not found" error: an unknown domain and a domain the account
  may not manage both return error 8 ("no permission"). The data source reports
  this as "Domain not found or not manageable".
- API error messages are localised by mailbox.org (German on German accounts);
  the provider only ever acts on the numeric error code.
- Values of `domain.get` are returned as text, as the API returns them (for
  example `status` is `registriert`).

## Development

```sh
make bootstrap   # go mod tidy
make all         # format, lint, test, docs, validate
```

Unit tests need no credentials: they run against a fake JSON-RPC server
(`internal/testutils`) and require a `terraform` binary on `PATH`.
Releases are cut by pushing a `v*` tag; CI builds and signs them with GoReleaser.

Licensed under the MIT License.
