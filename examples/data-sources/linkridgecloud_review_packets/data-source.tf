data "linkridgecloud_review_packets" "service_token" {
  account_id  = "acct_local_qr_demo"
  packet_type = "service_token"
  status      = "blocked_pending_matthew_approval"
}
