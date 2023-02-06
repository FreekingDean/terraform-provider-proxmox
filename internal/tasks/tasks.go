package tasks

import (
	"context"
	"fmt"

	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/tasks"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Client struct {
	c *tasks.Client
}

func New(p tasks.HTTPClient) *Client {
	return &Client{
		c: tasks.New(p),
	}
}

func (t *Client) Wait(ctx context.Context, upid string, node string) diag.Diagnostics {
	diag := diag.Diagnostics{}
	req := tasks.ReadTaskStatusRequest{
		Node: node,
		Upid: upid,
	}
	var exit string
	for {
		resp, err := t.c.ReadTaskStatus(ctx, req)
		if err != nil {
			diag.AddError(
				fmt.Sprintf("Error waiting for task id %s on %s", upid, node),
				"An unexpected error occurred when getting the task. "+
					"Proxmox Task Error: "+err.Error(),
			)
			return diag
		}
		if resp.Status != "running" {
			exit = *resp.Exitstatus
			break
		}
	}
	if exit != "OK" {
		diag.AddError(
			"Error waiting for VM to be created",
			"An unexpected error occurred when getting the task. "+
				"Proxmox Task Error: received bad exit status: "+exit,
		)
		return diag
	}
	return diag
}
