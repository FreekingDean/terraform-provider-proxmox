package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/FreekingDean/terraform-provider-proxmox/proxmox"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name proxmox

func main() {
	providerserver.Serve(context.Background(), proxmox.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/freekingdean/proxmox",
	})
}
