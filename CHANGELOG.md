# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Terraform Plugin Framework provider scaffold for LinkRidge Cloud.
- `LINKRIDGE_CLOUD_BASE_URL` and `LINKRIDGE_CLOUD_API_TOKEN` configuration
  fallback.
- Read-only `linkridgecloud_services` data source for the `/v1/services`
  control-plane API.
- Read-only `linkridgecloud_accounts` data source for `/v1/accounts` draft
  tenant planning records.
- Read-only `linkridgecloud_account_memberships` data source for
  `/v1/accounts/{account_id}/memberships` role assignment evidence without
  user creation, invite delivery, external access, or role changes.
- Read-only `linkridgecloud_account_invites` data source for
  `/v1/accounts/{account_id}/invites` draft invite delivery evidence without
  sending invites, notifying customers, or granting external identity access.
- Read-only `linkridgecloud_account_services` data source for
  `/v1/accounts/{account_id}/services` service planning records.
- Read-only `linkridgecloud_activation_packets` data source for safe
  `/v1/accounts/{account_id}/services/{account_service_id}/activation-packets`
  review and provisioning rehearsal metadata without activation.
- Read-only `linkridgecloud_service_tokens` data source for safe
  `/v1/accounts/{account_id}/service-tokens` metadata without secret material.
- Read-only `linkridgecloud_qr_import_jobs` data source for safe
  `/v1/qr/workspaces/{workspace_id}/import-jobs` planning and review metadata
  without import execution.
- Read-only `linkridgecloud_qr_mutation_rehearsal_requirements` data source for
  safe
  `/v1/qr/workspaces/{workspace_id}/mutation-requests/{mutation_request_id}/execution-rehearsal-requirements`
  checklist and guardrail evidence without executing QR mutations, writing
  tenant QR records, recording billing usage, or enabling hosted redirects.
- Read-only `linkridgecloud_qr_workspaces` data source for `/v1/qr/workspaces`
  QR service planning records.
- Read-only `linkridgecloud_provisioning_runs` data source for
  `/v1/provisioning-runs` rehearsal evidence without provisioning execution.
- Read-only `linkridgecloud_billing_export_requests` data source for
  `/v1/accounts/{account_id}/services/{account_service_id}/billing-export-requests`
  review metadata without billing export.
- Read-only `linkridgecloud_support_cases` data source for
  `/v1/accounts/{account_id}/services/{account_service_id}/support-cases`
  handoff metadata without external tickets or customer notifications.
- Read-only `linkridgecloud_review_packets` data source for `/v1/review-packets`
  normalized approval-gate evidence without approving or executing effects.
- Read-only `linkridgecloud_operator_approvals` data source for
  `/v1/operator-approvals` decision evidence without approving, rejecting, or
  executing effects.
- Read-only `linkridgecloud_audit_events` data source for `/v1/audit-events`
  append-only evidence without replaying events or causing effects.

### Changed

- Trim provider base URL and API token values before client use, send a stable
  provider user-agent, and format LinkRidge problem+json API errors for clearer
  Terraform diagnostics.
- Updated beta banner to v1 stable

## [1.0.0] - 2026-03-01

### Added

- Makefile with all 7 language ecosystems (Python, Bash, Terraform, Ansible, Ruby, Go, JavaScript/TypeScript)
- `make init` / `make _init` config scaffolding target
- CI workflows: lint, format, test, security, scan, docs
- Pre-commit hooks for all supported languages (commented out by default)
- Agent instruction files (CLAUDE.md, AGENTS.md, .cursorrules, .opencode/agents.yaml)
- DevRail compliance badge in README
- Retrofit guide for adding DevRail to existing repositories
- `.devrail.yml` with all 7 languages listed (commented out)
- `.editorconfig`, `.gitignore`, `DEVELOPMENT.md`, `CHANGELOG.md`, `LICENSE`
