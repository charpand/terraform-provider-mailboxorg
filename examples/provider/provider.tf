terraform {
  required_providers {
    mailboxorg = {
      source = "charpand/mailboxorg"
    }
  }
}

# Credentials can also come from MAILBOXORG_USERNAME and MAILBOXORG_PASSWORD.
provider "mailboxorg" {
  username = var.mailboxorg_username
  password = var.mailboxorg_password
}

variable "mailboxorg_username" {
  type = string
}

variable "mailboxorg_password" {
  type      = string
  sensitive = true
}

# Several logins in one configuration (for example a reseller login and the
# login of an account it manages): give each one an alias and credentials of
# its own, then pick it per resource or data source with `provider`.
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

variable "reseller_username" {
  type = string
}

variable "reseller_password" {
  type      = string
  sensitive = true
}

variable "account_username" {
  type = string
}

variable "account_password" {
  type      = string
  sensitive = true
}
