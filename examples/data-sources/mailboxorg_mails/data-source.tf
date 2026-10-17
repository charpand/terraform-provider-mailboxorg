data "mailboxorg_mails" "example" {
  domain = "example.com"
}

output "addresses" {
  value = [for m in data.mailboxorg_mails.example.mails : m.mail]
}
