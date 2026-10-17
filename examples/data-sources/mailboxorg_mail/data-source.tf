data "mailboxorg_mail" "example" {
  mail = "user@example.com"
}

output "mailbox_plan" {
  value = data.mailboxorg_mail.example.plan
}
