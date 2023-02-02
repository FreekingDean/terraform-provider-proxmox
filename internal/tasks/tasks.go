package tasks

import (
	"context"
	"fmt"

	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/tasks"
)

type Client struct {
	c *tasks.Client
}

func New(p tasks.HTTPClient) *Client {
	return &Client{
		c: tasks.New(p),
	}
}

func (t *Client) Wait(ctx context.Context, upid string, node string) error {
	req := tasks.ReadTaskStatusRequest{
		Node: node,
		Upid: upid,
	}
	var exit string
	for {
		resp, err := t.c.ReadTaskStatus(ctx, req)
		if err != nil {
			return err
		}
		if resp.Status != "running" {
			exit = *resp.Exitstatus
			break
		}
	}
	if exit != "OK" {
		return fmt.Errorf("received bad exit status: %s", exit)
	}
	return nil
}
