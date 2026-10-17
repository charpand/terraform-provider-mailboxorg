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

## Multiple logins

A reseller manages accounts, and each account manages its own domains and
mailboxes, so you may need more than one login in a single configuration and
state. Use provider aliases, each with credentials of its own, and select one
per resource or data source with `provider`:

```hcl
provider "mailboxorg" {
  alias    = "reseller"
  username = var.reseller_username
  password = var.reseller_password
}

provider "mailboxorg" {
  alias    = "account"
  username = var.account_username
  password = var.account_password
}

data "mailboxorg_domain" "example" {
  provider = mailboxorg.account
  domain   = "example.com"
}
```

- `MAILBOXORG_USERNAME` and `MAILBOXORG_PASSWORD` only apply to a provider block
  that does not set `username`/`password` itself. With several logins, set the
  credentials per alias, for example from `sensitive` variables.
- Every alias is a separate configured provider with its own client and its own
  mailbox.org session, so the logins do not interfere with each other.
- A module that needs a non-default login declares `configuration_aliases` and
  receives the provider from its caller (`providers = { mailboxorg.account = mailboxorg.account }`).
- Terraform cannot create providers dynamically with `for_each`; every login
  needs its own provider block (OpenTofu 1.9+ supports `for_each` on providers).
- Passwords passed as variables can end up in plan files and state. Do not
  commit them in `.tfvars` files.

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
