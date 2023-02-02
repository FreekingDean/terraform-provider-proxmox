package proxmox

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/storage"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/storage/content"

	"github.com/FreekingDean/terraform-provider-proxmox/internal/tasks"
)

type IsoModel struct {
	Url               types.String `tfsdk:"url"`
	Checksum          types.String `tfsdk:"checksum"`
	ChecksumAlgorithm types.String `tfsdk:"checksum_algorithm"`
}

type resourceNodeStorageContentModel struct {
	Storage  types.String `tfsdk:"storage"`
	Filename types.String `tfsdk:"filename"`
	ID       types.String `tfsdk:"id"`
	Iso      *IsoModel    `tfsdk:"iso"`
}

type resourceNodeStorageContent struct {
	s *storage.Client
	t *tasks.Client
	c *content.Client
}

func (r *resourceNodeStorageContent) SetClient(p *proxmox.Client) {
	r.s = storage.New(p)
	r.t = tasks.New(p)
	r.c = content.New(p)
}

func (r *resourceNodeStorageContent) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_storage_content"
}

func (e *resourceNodeStorageContent) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"filename": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"iso": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"checksum": schema.StringAttribute{
						Optional: true,
					},
					"checksum_algorithm": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

// TODO: Add once more blocks added
//func (r *resourceNodeStorageContent) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
//	return []resource.ConfigValidator{
//		resourcevalidator.Conflicting(
//			path.MatchRoot("iso"),
//		),
//	}
//}

func (r *resourceNodeStorageContent) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceNodeStorageContentModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := &StorageID{}
	err := id.SScan(plan.Storage.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retreiving content information",
			"An unexpected error occurred when retreiving content information. "+
				"Proxmox API Error: "+err.Error(),
		)
	}
	format := ""
	if plan.Iso != nil {
		format = "iso"
		outStr, err := r.s.DownloadUrl(ctx, storage.DownloadUrlRequest{
			Content:  "iso",
			Filename: plan.Filename.ValueString(),
			Url:      plan.Iso.Url.ValueString(),
			Node:     id.Node,
			Storage:  id.Storage,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error downloading iso",
				"An unexpected error occurred when downloading ISO. "+
					"Proxmox Client Error: "+err.Error(),
			)
			return
		}
		err = r.t.Wait(ctx, outStr, id.Node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error downloading iso",
				"An unexpected error occurred when downloading ISO. "+
					"Proxmox Task Error: "+err.Error(),
			)
			return
		}
	}
	plan.ID = types.StringValue(
		fmt.Sprintf("%s:%s/%s", id.Storage, format, plan.Filename.ValueString()),
	)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceNodeStorageContent) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceNodeStorageContentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := &StorageID{}
	err := id.SScan(data.Storage.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing storage identifier",
			"An unexpected error occurred when parsing the storage identifier. "+
				"Error: "+err.Error(),
		)
		return
	}
	taskID, err := r.c.Delete(ctx, content.DeleteRequest{
		Node:    id.Node,
		Volume:  data.ID.ValueString(),
		Storage: &id.Storage,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error downloading iso",
			"An unexpected error occurred when downloading ISO. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
	err = r.t.Wait(ctx, taskID, id.Node)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting content",
			"An unexpected error occurred when deleting content. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
}

func (r *resourceNodeStorageContent) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceNodeStorageContentModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceNodeStorageContent) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceNodeStorageContentModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := &StorageID{}
	err := id.SScan(state.Storage.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing storage identifier",
			"An unexpected error occurred when parsing the storage identifier. "+
				"Error: "+err.Error(),
		)
		return
	}
	_, err = r.c.Find(ctx, content.FindRequest{
		Node:    id.Node,
		Volume:  state.ID.ValueString(),
		Storage: &id.Storage,
	})
	if err != nil && !strings.Contains(err.Error(), "volume_size_info on") {
		resp.Diagnostics.AddError(
			"Error retreiving content information",
			"An unexpected error occurred when retreiving content information. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	} else if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceNodeStorageContent) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: node@volume:format/filename, got: %q", req.ID),
		)
		return
	}
	node := parts[0]
	volID := parts[1]
	parts = strings.Split(parts[1], ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: node@volume:format/filename, got: %q", req.ID),
		)
		return
	}
	storage := parts[0]
	parts = strings.Split(parts[1], "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: node@volume:format/filename, got: %q", req.ID),
		)
		return
	}
	format := parts[0]
	filename := parts[1]

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), volID)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("storage"), fmt.Sprintf("%s/%s", node, storage))...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("filename"), &filename)...,
	)

	if format == "iso" {
		resp.Diagnostics.Append(
			resp.State.SetAttribute(ctx, path.Root("iso"), &IsoModel{})...,
		)
	}
}

type StorageID struct {
	Node    string
	Storage string
}

func (id *StorageID) SScan(storageID string) error {
	parts := strings.Split(storageID, "/")
	if len(parts) != 2 {
		return fmt.Errorf("Bad storage ID format")
	}
	id.Node = parts[0]
	id.Storage = parts[1]
	return nil
}
