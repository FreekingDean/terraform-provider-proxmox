# Download storage content
resource "proxmox_node_storage_content" "ubuntu_iso" {
  node     = "node_one"
  storage  = "ds9/local"
  filename = "ubuntu-22.iso"

  iso {
    url               = "https://releases.ubuntu.com/22.04.1/ubuntu-22.04.1-live-server-amd64.iso"
    checksum          = "10f19c5b2b8d6db711582e0e27f5116296c34fe4b313ba45f9b201a5007056cb"
    checksum_alorithm = "sha256"
  }
}
