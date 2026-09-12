# Terraform Provider for LinkRidge Cloud

> Built with [DevRail](https://devrail.dev) `v1` standards. See [STABILITY.md](STABILITY.md) for component status.

Terraform provider for the LinkRidge Cloud control-plane API.

The first provider slice is intentionally read-only: it configures a
Terraform Plugin Framework provider, creates a small LinkRidge Cloud API
client, and exposes the `/v1/services`, `/v1/accounts`, and
`/v1/accounts/{account_id}/services`,
`/v1/accounts/{account_id}/services/{account_service_id}/activation-packets`,
`/v1/accounts/{account_id}/services/{account_service_id}/billing-export-requests`,
`/v1/accounts/{account_id}/services/{account_service_id}/support-cases`,
`/v1/accounts/{account_id}/service-tokens`, `/v1/provisioning-runs`,
`/v1/review-packets`,
`/v1/qr/workspaces`, and
`/v1/qr/workspaces/{workspace_id}/import-jobs` list APIs through Terraform data
sources. Activation packets and provisioning runs expose review/rehearsal
evidence only; service-token data is safe metadata only; QR import jobs expose
plan/review status only; billing export, support case, and review packet data
sources expose local operator evidence only. The provider never returns token
secret material, approves activation/provisioning, executes imports, creates
billing records, sends customer notifications, opens external support tickets,
or enables hosted redirects. Draft resources will stay plan/dev-only until the
LinkRidge Cloud API approval gates, tenant isolation, and durable audit contracts
are ready for customer-visible effects.

<!-- badges-start -->
[![DevRail compliant](https://devrail.dev/images/badge.svg)](https://devrail.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
<!-- badges-end -->

## Quick Start

Configure the provider with explicit attributes or environment variables:

```hcl
provider "linkridgecloud" {
  base_url  = "https://dev.cloud.linkridge.net"
  api_token = var.linkridge_cloud_api_token
}

data "linkridgecloud_services" "qr" {
  id = "qr-codes"
}

data "linkridgecloud_accounts" "draft" {
  status = "draft"
}

data "linkridgecloud_account_services" "planned" {
  account_id = "acct_local_qr_demo"
  status     = "planned"
}

data "linkridgecloud_activation_packets" "planned" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked_pending_matthew_approval"
}

data "linkridgecloud_qr_workspaces" "planned" {
  account_id = "acct_local_qr_demo"
  status     = "planned"
}

data "linkridgecloud_qr_import_jobs" "planned" {
  workspace_id      = "qrw_local_demo"
  status            = "planned"
  import_performed  = false
}

data "linkridgecloud_provisioning_runs" "rehearsal" {
  account_service_id = "asvc_local_qr_demo"
  status             = "rehearsal_only"
}

data "linkridgecloud_billing_export_requests" "blocked" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked"
}

data "linkridgecloud_support_cases" "blocked_import_review" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  category           = "import_review"
  status             = "blocked"
}

data "linkridgecloud_review_packets" "service_token" {
  account_id  = "acct_local_qr_demo"
  packet_type = "service_token"
  status      = "blocked_pending_matthew_approval"
}

data "linkridgecloud_service_tokens" "planned" {
  account_id             = "acct_local_qr_demo"
  status                 = "planned"
  secret_material_issued = false
}
```

Environment variable fallback:

```shell
export LINKRIDGE_CLOUD_BASE_URL="https://dev.cloud.linkridge.net"
export LINKRIDGE_CLOUD_API_TOKEN="..."
```

Never commit token values. The provider only sends the token as a bearer
credential to the configured LinkRidge Cloud API.

## Usage

The Makefile is the universal execution interface. Every target produces consistent behavior whether invoked by a developer, CI pipeline, or AI agent.

| Target | Purpose |
|---|---|
| `make help` | Show available targets (default) |
| `make lint` | Run all linters for declared languages |
| `make format` | Run all formatters for declared languages |
| `make fix` | Auto-fix formatting issues in-place |
| `make test` | Run project test suite |
| `make security` | Run language-specific security scanners |
| `make scan` | Run universal scanning (trivy, gitleaks) |
| `make docs` | Generate documentation |
| `make check` | Run all of the above; report composite summary |
| `make install-hooks` | Install pre-commit and pre-push hooks |

All targets except `help` and `install-hooks` delegate to the dev-toolchain Docker container (`ghcr.io/devrail-dev/dev-toolchain:v1`).

## Configuration

### `.devrail.yml`

Every DevRail-managed repository includes a `.devrail.yml` file at the repo root. This file declares the project's languages and settings, and is read by the Makefile, CI pipelines, and AI agents.

This repository currently enables the Go DevRail checks. Terraform examples are
kept under `examples/`; Terraform-specific DevRail checks can be enabled once
the shared DevRail toolchain image includes the matching Terraform security
scanner.

## Contributing

See [DEVELOPMENT.md](DEVELOPMENT.md) for development standards, coding conventions, and contribution guidelines.

This project follows [Conventional Commits](https://www.conventionalcommits.org/). All commits use the `type(scope): description` format.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
