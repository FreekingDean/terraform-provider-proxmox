package proxmox

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/qemu"

	"github.com/FreekingDean/terraform-provider-proxmox/internal/tasks"
)

// Type scsi,ide
// Media cdrom,disk
type Disk struct {
	//Type    types.String `tfsdk:"type"`
	//From    types.String `tfsdk:"from"`
	//Media   types.String `tfsdk:"media"`
	NodeStorage types.String `tfsdk:"node_storage"`
	SizeGB      types.Int64  `tfsdk:"size_gb"`
	Content     types.String `tfsdk:"content"`
}

type Network struct {
	Bridge     types.String `tfsdk:"bridge"`
	Firewall   types.Bool   `tfsdk:"firewall"`
	MacAddress types.String `tfsdk:"mac_address"`
}

type resourceNodeVirtualMachineModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Node     types.String `tfsdk:"node"`
	Ides     []*Disk      `tfsdk:"ide"`
	Scsis    []*Disk      `tfsdk:"scsi"`
	Networks []*Network   `tfsdk:"network"`
	//Disks []*Disk      `tfsdk:"storage"`
	Memory types.Int64 `tfsdk:"memory"`
	CPUs   types.Int64 `tfsdk:"cpus"`
}

type resourceNodeVirtualMachine struct {
	t *tasks.Client
	q *qemu.Client
}

func (r *resourceNodeVirtualMachine) SetClient(p *proxmox.Client) {
	r.t = tasks.New(p)
	r.q = qemu.New(p)
}

func (r *resourceNodeVirtualMachine) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_virtual_machine"
}

func (e *resourceNodeVirtualMachine) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	diskBlock := schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"content": schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"size_gb": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.UseStateForUnknown(),
					},
				},
				"node_storage": schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory": schema.Int64Attribute{
				Required: true,
			},
			"cpus": schema.Int64Attribute{
				Required: true,
			},
			"node": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ide":  diskBlock,
			"scsi": diskBlock,
			"network": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bridge": schema.StringAttribute{
							Required: true,
						},
						"firewall": schema.BoolAttribute{
							Required: true,
						},
						"mac_address": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceNodeVirtualMachine) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceNodeVirtualMachineModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mem := int(plan.Memory.ValueInt64())
	cores := int(plan.CPUs.ValueInt64())
	creq := qemu.CreateRequest{
		Node:   plan.Node.ValueString(),
		Vmid:   int(plan.ID.ValueInt64()),
		Memory: &mem,
		Cores:  &cores,
	}
	ideArr := make(qemu.Ides, 0)
	scsiArr := make(qemu.Scsis, 0)

	for _, d := range plan.Ides {
		ide := &qemu.Ide{}
		if !d.Content.IsNull() {
			ide.File = d.Content.ValueString()
			if filepath.Ext(d.Content.ValueString()) == ".iso" {
				ide.Media = qemu.PtrIdeMedia(qemu.IdeMedia_CDROM)
			}
		} else if !d.NodeStorage.IsNull() {
			ide.File = fmt.Sprintf("%s:%d", d.NodeStorage.ValueString(), d.SizeGB.ValueInt64())
		}
		ideArr = append(ideArr, ide)
	}

	for _, d := range plan.Scsis {
		scsi := &qemu.Scsi{}
		if !d.Content.IsNull() {
			scsi.File = d.Content.ValueString()
			if filepath.Ext(d.Content.ValueString()) == ".iso" {
				scsi.Media = qemu.PtrScsiMedia(qemu.ScsiMedia_CDROM)
			}
		} else if !d.NodeStorage.IsNull() {
			scsi.File = fmt.Sprintf("%s:%d", d.NodeStorage.ValueString(), d.SizeGB.ValueInt64())
		}
		scsiArr = append(scsiArr, scsi)
	}
	creq.Ides = &ideArr
	creq.Scsis = &scsiArr
	task, err := r.q.Create(ctx, creq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VM",
			"An unexpected error occurred when creating the VM. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}

	err = r.t.Wait(ctx, task, plan.Node.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for VM to be created",
			"An unexpected error occurred when waiting for the VM. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceNodeVirtualMachine) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceNodeVirtualMachineModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	taskID, err := r.q.Delete(ctx, qemu.DeleteRequest{
		Node:  data.Node.ValueString(),
		Vmid:  int(data.ID.ValueInt64()),
		Purge: proxmox.PVEBool(true),
	})
	if err.Error() == fmt.Sprintf("non 200: 500 Configuration file 'nodes/%s/qemu-server/%d.conf' does not exist", data.Node.ValueString(), data.ID.ValueInt64()) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting VM",
			"An unexpected error occurred when deleting the VM. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}
	err = r.t.Wait(ctx, taskID, data.Node.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for VM to be created",
			"An unexpected error occurred when waiting for the VM. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
}

func (r *resourceNodeVirtualMachine) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceNodeVirtualMachineModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configReq := qemu.UpdateVmAsyncConfigRequest{
		Node: plan.Node.ValueString(),
		Vmid: int(plan.ID.ValueInt64()),
	}
	configReq.Memory = proxmox.Int(int(plan.Memory.ValueInt64()))
	configReq.Cores = proxmox.Int(int(plan.CPUs.ValueInt64()))
	ideArr := make(qemu.Ides, len(plan.Ides))
	for i, d := range plan.Ides {
		ide := &qemu.Ide{}
		if !d.Content.IsNull() {
			ide.File = d.Content.ValueString()
			if filepath.Ext(d.Content.ValueString()) == ".iso" {
				ide.Media = qemu.PtrIdeMedia(qemu.IdeMedia_CDROM)
			}
			plan.Ides[i].SizeGB = types.Int64Value(0)
			plan.Ides[i].NodeStorage = types.StringValue(
				strings.Split(d.Content.ValueString(), ":")[0],
			)
		} else if !d.NodeStorage.IsNull() {
			ide.File = fmt.Sprintf("%s:%d", d.NodeStorage.ValueString(), d.SizeGB.ValueInt64())
		}
		ideArr[i] = ide
	}

	scsiArr := make(qemu.Scsis, len(plan.Scsis))
	for i, d := range plan.Scsis {
		scsi := &qemu.Scsi{}
		if !d.Content.IsNull() {
			scsi.File = d.Content.ValueString()
			if filepath.Ext(d.Content.ValueString()) == ".iso" {
				scsi.Media = qemu.PtrScsiMedia(qemu.ScsiMedia_CDROM)
			}
		} else if !d.NodeStorage.IsNull() {
			scsi.File = fmt.Sprintf("%s:%d", d.NodeStorage.ValueString(), d.SizeGB.ValueInt64())
		}
		scsiArr[i] = scsi
	}
	configReq.Ides = &ideArr
	configReq.Scsis = &scsiArr

	nets := make(qemu.Nets, len(plan.Networks))
	for i, net := range plan.Networks {
		n := &qemu.Net{
			Firewall: proxmox.PVEBool(net.Firewall.ValueBool()),
			Bridge:   proxmox.String(net.Bridge.ValueString()),
			Model:    qemu.NetModel_VIRTIO,
		}
		if net.MacAddress.IsNull() || net.MacAddress.ValueString() == "" {
			mac, err := generateMac()
			if err != nil {
				resp.Diagnostics.AddError(
					"Error generating mac",
					"An unexpected error occurred when generating a mac address. "+
						"generateMac Error: "+err.Error(),
				)
				return
			}
			n.Macaddr = proxmox.String(mac.String())
			net.MacAddress = types.StringValue(mac.String())
		} else {
			n.Macaddr = proxmox.String(net.MacAddress.ValueString())
		}
		nets[i] = n
	}
	configReq.Nets = &nets

	task, err := r.q.UpdateVmAsyncConfig(ctx, configReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating VM",
			"An unexpected error occurred when updating the VM. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}

	err = r.t.Wait(ctx, task, plan.Node.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for VM to be created",
			"An unexpected error occurred when waiting for the VM. "+
				"Proxmox Task Error: "+err.Error(),
		)
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceNodeVirtualMachine) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceNodeVirtualMachineModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.q.VmConfig(ctx, qemu.VmConfigRequest{
		Node: state.Node.ValueString(),
		Vmid: int(state.ID.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error gettng  VM config",
			"An unexpected error occurred when retreiving the VM config. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}
	if config.Memory == nil || config.Cores == nil {
		resp.Diagnostics.AddError(
			"Memory or Cpus nil",
			"An unexpected error occurred when retreiving the VM Mem & CPU config.",
		)
		return
	}
	state.Memory = types.Int64Value(int64(*config.Memory))
	state.CPUs = types.Int64Value(int64(*config.Cores))
	state.Ides = make([]*Disk, 0)
	if config.Ides != nil {
		for i, ide := range *config.Ides {
			if ide == nil {
				state.Ides = append(state.Ides, nil)
				continue
			}
			disk, diags := buildDisk(i, ide.File, ide.Size)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.Ides = append(state.Ides, disk)
		}
	}
	state.Scsis = make([]*Disk, 0)
	if config.Scsis != nil {
		for i, scsi := range *config.Scsis {
			disk, diags := buildDisk(i, scsi.File, scsi.Size)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.Scsis = append(state.Scsis, disk)
		}
	}

	if config.Nets != nil {
		state.Networks = make([]*Network, len(*config.Nets))
		for i, net := range *config.Nets {
			n := &Network{}
			if net.Firewall != nil {
				n.Firewall = types.BoolValue(bool(*net.Firewall))
			}
			if net.Macaddr != nil {
				n.MacAddress = types.StringValue(*net.Macaddr)
			}
			if net.Bridge != nil {
				n.Bridge = types.StringValue(*net.Bridge)
			}
			state.Networks[i] = n
		}
	} else {
		state.Networks = make([]*Network, 0)
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//func (r *resourceNodeVirtualMachine) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
//	return []resource.ConfigValidator{
//		resourcevalidator.Conflicting(
//			path.MatchRoot("disk").AtName("content"),
//			path.MatchRoot("disk").AtName("node_storage"),
//		),
//		resourcevalidator.RequiredTogether(
//			path.MatchRoot("disk").AtName("size_gb"),
//			path.MatchRoot("disk").AtName("node_storage"),
//		),
//	}
//}

func buildDisk(i int, file string, sizeStr *string) (*Disk, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	storageID := strings.Split(file, ":")[0]
	var size int
	if sizeStr != nil {
		var err error
		size, err = strToGB(*sizeStr)
		if err != nil {
			diags.AddError(
				"Error converting size",
				fmt.Sprintf(
					"An unexpected error occurred when converting disk(ide%d) size(%s). ",
					i, *sizeStr,
				)+"Proxmox API Error: "+err.Error(),
			)
			return nil, diags
		}
	}
	return &Disk{
		NodeStorage: types.StringValue(storageID),
		Content:     types.StringValue(file),
		SizeGB:      types.Int64Value(int64(size)),
	}, diags
}
