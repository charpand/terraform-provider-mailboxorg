data "mailboxorg_domain" "example" {
  domain = "example.com"
}

output "context_id" {
  value = data.mailboxorg_domain.example.context_id
}
