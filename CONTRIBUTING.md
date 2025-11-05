# Contributing to terraform-provider-n8n

Thank you for your interest in contributing to the n8n Terraform provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions. We're all here to build something great together.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Your environment (OS, Terraform version, n8n version)
6. Relevant logs or error messages

### Suggesting Enhancements

We welcome feature requests! Please open an issue with:

1. A clear description of the feature
2. Use cases and examples
3. Why this would be useful to other users
4. Any implementation ideas you have

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the coding standards below
3. **Add tests** if applicable
4. **Update documentation** as needed
5. **Ensure tests pass** by running `go test ./...`
6. **Format your code** with `go fmt ./...`
7. **Run code generation** with `go generate ./...` if you modified provider code
8. **Add a changelog label** - Add one of the `changelog/*` labels to your PR:
   - `changelog/feature` - New features or enhancements
   - `changelog/bug` - Bug fixes
   - `changelog/improvement` - Improvements to existing features
   - `changelog/breaking-change` - Breaking changes
   - `changelog/documentation` - Documentation updates
   - `changelog/dependency` - Dependency updates
   - `changelog/no-changelog` - Changes that don't require changelog entry (CI, tests, etc.)
9. **Use proper PR title format** - Format: `[resource_name] Description` (e.g., `[n8n_workflow] Add tags support`)
   - Not required for PRs with `changelog/note` or `changelog/no-changelog` labels
10. **Submit a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.21 or later
- Terraform 1.0 or later
- Access to an n8n instance for testing

### Setting Up Your Development Environment

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/terraform-provider-n8n
cd terraform-provider-n8n

# Install dependencies
go mod download

# Build the provider
go build -o terraform-provider-n8n

# Run tests
go test -v ./...
```

### Local Testing

To test your changes locally:

```bash
# Build the provider
make build

# Install it locally
make install

# Create a test configuration in a separate directory
mkdir test-config
cd test-config

# Create a test main.tf file
cat > main.tf <<EOF
terraform {
  required_providers {
    n8n = {
      source = "pinotelio/n8n"
    }
  }
}

provider "n8n" {
  endpoint = "https://your-n8n-instance.com"
  api_key  = "your-api-key"
}

resource "n8n_workflow" "test" {
  name   = "Test Workflow"
  active = false
  nodes  = jsonencode([...])
  connections = jsonencode({...})
}
EOF

# Test it
terraform init
terraform plan
terraform apply
```

## Coding Standards

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small

### Code Organization

```
terraform-provider-n8n/
├── internal/
│   ├── client/          # API client implementation
│   │   └── client.go
│   └── provider/        # Provider implementation
│       ├── provider.go
│       ├── *_resource.go
│       └── *_data_source.go
├── examples/            # Example configurations
├── docs/               # Documentation
└── main.go            # Entry point
```

### Naming Conventions

- **Files**: Use snake_case (e.g., `workflow_resource.go`)
- **Types**: Use PascalCase (e.g., `WorkflowResource`)
- **Functions**: Use camelCase for private, PascalCase for exported
- **Variables**: Use camelCase

### Error Handling

Always provide clear, actionable error messages:

```go
// Good
if err != nil {
    resp.Diagnostics.AddError(
        "Error Creating Workflow",
        fmt.Sprintf("Could not create workflow: %s", err.Error()),
    )
    return
}

// Bad
if err != nil {
    resp.Diagnostics.AddError("Error", err.Error())
    return
}
```

### Testing

Write tests for new functionality:

```go
func TestWorkflowResource_Create(t *testing.T) {
    // Test implementation
}
```

Run tests:

```bash
go test -v ./...
```

## Adding New Resources

To add a new resource (e.g., `n8n_tag`):

1. **Create the API client methods** in `internal/client/client.go`:

```go
type Tag struct {
    ID   string `json:"id,omitempty"`
    Name string `json:"name"`
}

func (c *Client) CreateTag(tag *Tag) (*Tag, error) {
    // Implementation
}

func (c *Client) GetTag(id string) (*Tag, error) {
    // Implementation
}

func (c *Client) UpdateTag(id string, tag *Tag) (*Tag, error) {
    // Implementation
}

func (c *Client) DeleteTag(id string) error {
    // Implementation
}
```

2. **Create the resource file** `internal/provider/tag_resource.go`:

```go
package provider

import (
    "context"
    // imports...
)

type tagResource struct {
    client *client.Client
}

type tagResourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}

// Implement resource.Resource interface
func (r *tagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    // Define schema
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Implement Create
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implement Read
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Implement Update
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Implement Delete
}
```

3. **Register the resource** in `internal/provider/provider.go`:

```go
func (p *n8nProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewWorkflowResource,
        NewCredentialResource,
        NewTagResource,  // Add this
    }
}
```

4. **Add examples** in `examples/resources/n8n_tag/resource.tf`

5. **Update documentation** in `README.md`

## Adding New Data Sources

Similar to resources, but implement `datasource.DataSource` interface instead.

## Documentation

### Code Comments

- Add comments for all exported types and functions
- Use complete sentences
- Explain the "why" not just the "what"

### README Updates

When adding new features, update:

- Feature list
- Usage examples
- API reference

### Example Configurations

Add practical examples in the `examples/` directory.

## Commit Messages

Write clear, descriptive commit messages:

```
Add support for n8n tags

- Implement tag resource with CRUD operations
- Add tag data source
- Include examples and documentation
- Add tests for tag operations
```

Format:
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description with bullet points

## Pull Request Process

1. **Update documentation** for any user-facing changes
2. **Add tests** for new functionality
3. **Add changelog label** to your PR
4. **Use proper PR title format** (see above)
5. **Ensure CI passes** - All GitHub Actions workflows must pass:
   - Tests workflow (linting, build, tests)
   - Documentation workflow (if docs changed)
   - Changelog workflow (label and title validation)
6. **Request review** from maintainers
7. **Address feedback** promptly
8. **Squash commits** if requested

## Release Process

We follow a PR-based release process:

### For Maintainers

1. **Prepare Release:**
   - Go to Actions → "Prepare release" workflow
   - Click "Run workflow"
   - Enter version number (e.g., `0.2.0` without `v` prefix)
   - This creates a release PR with updated CHANGELOG.md

2. **Review Release PR:**
   - Review the generated CHANGELOG.md
   - Make any necessary adjustments
   - Ensure all changes are properly documented

3. **Merge Release PR:**
   - Merge the PR to main
   - The release workflow automatically:
     - Creates a git tag (e.g., `v0.2.0`)
     - Builds release artifacts with GoReleaser
     - Creates a GitHub release
     - Signs artifacts with GPG

4. **Verify Release:**
   - Check that the tag was created
   - Verify the GitHub release is published
   - Confirm artifacts are available and signed

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes (breaking changes)
- **MINOR** version for new functionality in a backwards compatible manner
- **PATCH** version for backwards compatible bug fixes

### Hotfix Process

For urgent fixes:

1. Create a hotfix branch from main
2. Make the fix and open a PR with `changelog/bug` label
3. After merging, run the prepare_release workflow
4. Follow standard release process

## Questions?

Feel free to:

- Open an issue for questions
- Start a discussion
- Reach out to maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉

