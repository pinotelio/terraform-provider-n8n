# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-04

### Added

- Initial release of the n8n Terraform provider
- Provider configuration with `endpoint` and `api_key` (supports environment variables)
- **Resources:**
  - `n8n_workflow` - Manage workflows with full CRUD operations
  - `n8n_credential` - Manage credentials with sensitive data handling
  - `n8n_user` - Manage users with role-based access control
  - `n8n_workflow_activation` - Control workflow activation state
- **Data Sources:**
  - `n8n_workflow` - Read workflow information
  - `n8n_user` - Read user information
- Import support for existing resources
- Comprehensive documentation and examples

### Notes

- Credentials use state-only management due to n8n API limitations (no read endpoint available)
- No drift detection for credentials

[0.1.0]: https://github.com/pinotelio/terraform-provider-n8n/releases/tag/v0.1.0

