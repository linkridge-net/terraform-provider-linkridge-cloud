data "linkridgecloud_activation_packets" "planned" {
  account_id         = "acct_local_qr_demo"
  account_service_id = "asvc_local_qr_demo"
  status             = "blocked_pending_matthew_approval"
}
