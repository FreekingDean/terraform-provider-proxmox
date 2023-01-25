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

data "proxmox_access_role" "firewall" {
  roleid = "PVEDatastoreAdmin"
}

output "value" {
  value = data.proxmox_access_role.firewall
}
