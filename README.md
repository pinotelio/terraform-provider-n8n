# Terraform Provider for n8n

[![Tests](https://github.com/pinotelio/terraform-provider-n8n/actions/workflows/test.yml/badge.svg)](https://github.com/pinotelio/terraform-provider-n8n/actions/workflows/test.yml)
[![Release](https://github.com/pinotelio/terraform-provider-n8n/actions/workflows/release.yml/badge.svg)](https://github.com/pinotelio/terraform-provider-n8n/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4)](https://registry.terraform.io/providers/pinotelio/n8n)

Manage your [n8n](https://n8n.io/) workflows, credentials, and users as code with Terraform. This provider enables Infrastructure as Code (IaC) practices for your n8n automation platform.

## Features

- **Workflow Management**: Full CRUD operations for n8n workflows
  - Import workflows from JSON files or paste directly
  - Support for complex workflows with multiple nodes and connections
  - Workflow activation control
- **Credential Management**: Securely manage n8n credentials
  - Support for all n8n credential types
  - Sensitive data protection
- **User Management**: Manage n8n users and roles
  - Support for owner, admin, and member roles
  - Email-based user management
- **Data Sources**: Query existing workflows and users
- **Import Support**: Import existing resources into Terraform state

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building the provider)
- n8n self-hosted instance with API access enabled
- n8n API key with appropriate permissions

## Quick Start

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    n8n = {
      source  = "pinotelio/n8n"
      version = "~> 0.1"
    }
  }
}

provider "n8n" {
  endpoint = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key  # Use variables for sensitive data
}
```

Run `terraform init` to download the provider automatically from the [Terraform Registry](https://registry.terraform.io/providers/pinotelio/n8n).

### Getting Your n8n API Key

1. Log in to your n8n instance
2. Go to **Settings** → **API**
3. Create a new API key
4. Use it in your provider configuration or set the `N8N_API_KEY` environment variable

### Environment Variables (Recommended)

For better security, use environment variables instead of hardcoding credentials:

```bash
export N8N_ENDPOINT="https://your-n8n-instance.com"
export N8N_API_KEY="your-api-key-here"
terraform plan
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Reporting bugs and requesting features
- Setting up your development environment
- Code style and testing requirements
- Submitting pull requests

## License

This provider is distributed under the MIT License. See `LICENSE` for more information.

## Support

- **Issues & Bug Reports**: [GitHub Issues](https://github.com/pinotelio/terraform-provider-n8n/issues)
- **n8n Documentation**: [docs.n8n.io](https://docs.n8n.io/)
- **n8n API Reference**: [API Documentation](https://docs.n8n.io/api/api-reference/)

## Acknowledgments

Built with the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) for the [n8n](https://n8n.io/) workflow automation platform.

