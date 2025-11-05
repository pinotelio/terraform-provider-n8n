data "n8n_user" "example" {
  id = "1"
}

output "user_email" {
  value = data.n8n_user.example.email
}

output "user_role" {
  value = data.n8n_user.example.role
}

output "user_id" {
  value = data.n8n_user.example.id
}

output "is_owner" {
  value = data.n8n_user.example.is_owner
}

