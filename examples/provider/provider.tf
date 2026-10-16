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
