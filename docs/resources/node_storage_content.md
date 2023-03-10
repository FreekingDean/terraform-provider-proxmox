---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "proxmox_node_storage_content Resource - proxmox"
subcategory: ""
description: |-
  Storage content aka a volume
---

# proxmox_node_storage_content (Resource)

Storage content aka a volume

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `filename` (String) The filename on the storage
- `storage` (String) The storage identifier

### Optional

- `iso` (Block, Optional) An iso object (see [below for nested schema](#nestedblock--iso))

### Read-Only

- `id` (String) The volid of the content

<a id="nestedblock--iso"></a>
### Nested Schema for `iso`

Required:

- `url` (String) The url to download the iso from

Optional:

- `checksum` (String) A checksum of the downlaoded content
- `checksum_algorithm` (String) The checksum algorithm of the downlaoded content

## Import

Import is supported using the following syntax:

```shell
# Order can be imported by specifying the numeric identifier.
terraform import proxmox_node_storage_content.ubuntu_iso node_one@local:iso/ubuntu.iso
```
