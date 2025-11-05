terraform {
  required_providers {
    n8n = {
      source = "pinotelio/n8n"
    }
  }
}

provider "n8n" {
  endpoint = "https://your-n8n-instance.com"
  api_key  = "your-api-key-here"
}

# Alternatively, you can use environment variables:
# export N8N_ENDPOINT="https://your-n8n-instance.com"
# export N8N_API_KEY="your-api-key-here"

