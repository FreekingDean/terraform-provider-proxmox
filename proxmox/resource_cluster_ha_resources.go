package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/ha/resources"
)

type resourceClusterHAResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Comment     types.String `tfsdk:"comment"`
	Group       types.String `tfsdk:"group"`
	MaxRelocate types.Int64  `tfsdk:"max_relocate"`
	MaxRestart  types.Int64  `tfsdk:"max_restart"`
	State       types.String `tfsdk:"state"`
}

type resourceClusterHAResource struct {
	r *resources.Client
}

func (r *resourceClusterHAResource) SetClient(p *proxmox.Client) {
	r.r = resources.New(p)
}

func (r *resourceClusterHAResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_ha_resources"
}

func (e *resourceClusterHAResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cluster HA Resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The resource ID (vm:101)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "A helpful comment on the HA resource",
			},
			"group": schema.StringAttribute{
				Optional:    true,
				Description: "The HA Group Identifier",
			},
			"max_relocate": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of times to relocate the resource (default: 1)",
			},
			"max_restart": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum number of times to restart the resource (default: 1)",
			},
			"state": schema.StringAttribute{
				Optional:    true,
				Description: "The desired state of the resource (default: started)",
			},
		},
	}
}

func (r *resourceClusterHAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceClusterHAResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.r.Create(ctx, resources.CreateRequest{
		Sid:         plan.ID.ValueString(),
		Comment:     proxmox.String(plan.Comment.ValueString()),
		MaxRestart:  proxmox.Int(int(plan.MaxRestart.ValueInt64())),
		MaxRelocate: proxmox.Int(int(plan.MaxRelocate.ValueInt64())),
		State:       resources.PtrState(resources.State(plan.State.ValueString())),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HA resource",
			"An unexpected error occurred when creating the HA resource. "+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceClusterHAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceClusterHAResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.r.Delete(ctx, resources.DeleteRequest{
		Sid: data.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ha resource",
			"An unexpected error occurred when deleting the ha resource. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
}

func (r *resourceClusterHAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceClusterHAResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.r.Update(ctx, resources.UpdateRequest{
		Sid:         plan.ID.ValueString(),
		Comment:     proxmox.String(plan.Comment.ValueString()),
		MaxRestart:  proxmox.Int(int(plan.MaxRestart.ValueInt64())),
		MaxRelocate: proxmox.Int(int(plan.MaxRelocate.ValueInt64())),
		State:       resources.PtrState(resources.State(plan.State.ValueString())),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HA resource",
			"An unexpected error occurred when creating the HA resource. "+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceClusterHAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceClusterHAResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.r.Find(ctx, resources.FindRequest{
		Sid: state.ID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HA resource",
			"An unexpected error occurred when reading the HA resource. "+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}
	if res.Comment != nil {
		state.Comment = types.StringValue(*res.Comment)
	}
	if res.Group != nil {
		state.Group = types.StringValue(*res.Group)
	}
	if res.State != nil {
		state.State = types.StringValue(string(*res.State))
	}
	if res.MaxRelocate != nil {
		state.MaxRelocate = types.Int64Value(int64(*res.MaxRelocate))
	}
	if res.MaxRestart != nil {
		state.MaxRestart = types.Int64Value(int64(*res.MaxRestart))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceClusterHAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	res, err := r.r.Find(ctx, resources.FindRequest{
		Sid: req.ID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HA resource",
			"An unexpected error occurred when reading the HA resource. "+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	if res.Comment != nil {
		resp.State.SetAttribute(ctx, path.Root("comment"), *res.Comment)
	}
	if res.Group != nil {
		resp.State.SetAttribute(ctx, path.Root("group"), *res.Group)
	}
	if res.State != nil {
		resp.State.SetAttribute(ctx, path.Root("state"), *res.State)
	}
	if res.MaxRelocate != nil {
		resp.State.SetAttribute(ctx, path.Root("max_relocate"), *res.MaxRelocate)
	}
	if res.MaxRestart != nil {
		resp.State.SetAttribute(ctx, path.Root("max_restart"), *res.MaxRestart)
	}
}
