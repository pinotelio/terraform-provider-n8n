# Create a basic user
resource "n8n_user" "member" {
  email = "user@example.com"
  role  = "global:member"
}

# Create an admin user
resource "n8n_user" "admin" {
  email = "admin@example.com"
  role  = "global:admin"
}

# Create a user with minimal information (email only, defaults to global:member role)
resource "n8n_user" "minimal" {
  email = "minimal@example.com"
}

# Create multiple users using for_each
variable "team_members" {
  type = map(object({
    email = string
    role  = string
  }))
  default = {
    "dev1" = {
      email = "dev1@example.com"
      role  = "global:member"
    }
    "dev2" = {
      email = "dev2@example.com"
      role  = "global:member"
    }
  }
}

resource "n8n_user" "team" {
  for_each = var.team_members

  email = each.value.email
  role  = each.value.role
}

