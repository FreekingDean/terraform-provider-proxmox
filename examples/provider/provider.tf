terraform {
  required_providers {
    proxmox = {
      source = "registry.terraform.io/freekingdean/proxmox"
    }
  }
}

provider "proxmox" {
  #host     = "https://192.168.1.111:8006/api2/json"
  #username = "myuser@pve"
  #password = "somepass"
}

resource "proxmox_node_storage_content" "iso" {
  filename = "k3os-old.iso"
  storage  = "ds9/local"
  iso = {
    url = "https://github.com/rancher/k3os/releases/download/v0.22.2-k3s2r0/k3os-amd64.iso"
  }
}
