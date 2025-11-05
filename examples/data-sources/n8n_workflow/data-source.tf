data "n8n_workflow" "example" {
  id = "1"
}

output "workflow_name" {
  value = data.n8n_workflow.example.name
}

output "workflow_active" {
  value = data.n8n_workflow.example.active
}

