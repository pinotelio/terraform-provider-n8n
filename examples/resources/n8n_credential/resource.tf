resource "n8n_credential" "example" {
  name = "My API Credential"
  type = "httpBasicAuth"

  data = jsonencode({
    user     = "myusername"
    password = "mypassword"
  })
}

# Example with Datadog API credential
resource "n8n_credential" "datadog" {
  name = "Datadog Account"
  type = "datadogApi"

  data = jsonencode({
    apiKey = "xoxb-your-datadog-api-key"
    appKey = "xoxb-your-datadog-app-key"
  })
}

# Example with Datadog API credential using n8n environment variables
resource "n8n_credential" "datadog_env" {
  name = "Datadog Account"
  type = "datadogApi"

  data = jsonencode({
    apiKey = "{{ \\$.env.DATADOG_API_KEY }}"
    appKey = "{{ \\$.env.DATADOG_APP_KEY }}"
  })
}
