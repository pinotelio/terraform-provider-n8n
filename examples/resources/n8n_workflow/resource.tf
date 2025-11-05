# Example 1: Using individual attributes with proper trigger node
resource "n8n_workflow" "example" {
  name = "My Example Workflow"

  nodes = jsonencode([
    {
      id          = "1"
      name        = "Schedule Trigger"
      type        = "n8n-nodes-base.scheduleTrigger"
      typeVersion = 1.2
      position    = [250, 300]
      parameters = {
        rule = {
          interval = [
            {
              field         = "hours"
              hoursInterval = 1
            }
          ]
        }
      }
    },
    {
      id          = "2"
      name        = "HTTP Request"
      type        = "n8n-nodes-base.httpRequest"
      typeVersion = 4.2
      position    = [450, 300]
      parameters = {
        url    = "https://api.example.com/data"
        method = "GET"
      }
    }
  ])

  connections = jsonencode({
    "Schedule Trigger" = {
      main = [
        [
          {
            node  = "HTTP Request"
            type  = "main"
            index = 0
          }
        ]
      ]
    }
  })

  settings = jsonencode({
    executionOrder = "v1"
  })
}

# Example 2: Using workflow_json with file() function (recommended for complex workflows)
resource "n8n_workflow" "from_file" {
  workflow_json = file("${path.module}/workflows/some-workflow.json")
}

# Example 3: Using workflow_json with jsondecode for dynamic values
locals {
  workflow_template = jsondecode(file("${path.module}/workflows/some-template.json"))
}

resource "n8n_workflow" "dynamic" {
  workflow_json = jsonencode(merge(
    local.workflow_template,
    {
      name = "Dynamic Workflow - ${var.environment}"
    }
  ))
}
