data "linkridgecloud_audit_events" "service_token_prepare" {
  account_id       = "acct_local_qr_demo"
  action           = "service_token.prepare"
  source_packet_id = "local-qr-starter-demo"
}
