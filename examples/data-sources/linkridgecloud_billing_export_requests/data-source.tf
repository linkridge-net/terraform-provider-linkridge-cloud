data "linkridgecloud_billing_export_requests" "blocked" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked"
}
