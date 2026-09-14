data "linkridgecloud_provisioning_runs" "rehearsal" {
  account_service_id = "asvc_local_qr_demo"
  status             = "rehearsal_only"
}
