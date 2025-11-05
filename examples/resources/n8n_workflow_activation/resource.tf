resource "n8n_workflow_activation" "example" {
  workflow_id = n8n_workflow.example.id
  active      = true
}