# Create a virtual machine
resource "proxmox_node_virtual_machine" "ubuntu" {
  node = "node_one"

  memory = 2048 # 2 GB
  cpus   = 4    # 4 Cores

  ide {
    content = "local:iso/ubuntu.iso"
  }

  scsi {
    import_from = "local:999/template_disk.qcow2"
    read_only   = true
  }

  scsi {
    node_storage = "local"
    size_gb      = 100
  }

  network {
    bridge   = "vmbr0"
    firewall = true
  }
}
