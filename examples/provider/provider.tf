terraform {
  required_providers {
    proxmox = {
      source = "registry.terraform.io/freekingdean/proxmox"
    }

    null = {
      source  = "hashicorp/null"
      version = "3.2.1"
    }
  }
}

provider "null" {}

provider "proxmox" {
  #host     = "https://192.168.1.111:8006/api2/json"
  #username = "myuser@pve"
  #password = "somepass"
}
