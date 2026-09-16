data "linkridgecloud_operator_approvals" "blocked" {
  account_id        = "acct_local_qr_demo"
  decision_state    = "blocked"
  required_approval = "matthew"
}
