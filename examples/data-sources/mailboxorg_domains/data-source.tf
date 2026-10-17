data "mailboxorg_domains" "example" {
  account = "my-account"
}

output "domain_names" {
  value = [for d in data.mailboxorg_domains.example.domains : d.domain]
}
