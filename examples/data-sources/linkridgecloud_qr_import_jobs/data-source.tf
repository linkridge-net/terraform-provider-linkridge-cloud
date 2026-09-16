data "linkridgecloud_qr_import_jobs" "planned" {
  workspace_id      = "qrw_local_demo"
  status            = "planned"
  import_performed  = false
}
