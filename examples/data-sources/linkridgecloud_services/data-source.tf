provider "linkridgecloud" {
  base_url = "https://dev.cloud.linkridge.net"
}

data "linkridgecloud_services" "qr" {
  id = "qr-codes"
}
