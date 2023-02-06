package proxmox

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/qemu"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/qemu/status"

	"github.com/FreekingDean/terraform-provider-proxmox/internal/tasks"
	"github.com/FreekingDean/terraform-provider-proxmox/internal/validators"
)

// Type scsi,ide
// Media cdrom,disk
type Disk struct {
	ID         types.Int64  `tfsdk:"id"`
	VolumeID   types.String `tfsdk:"volume_id"`
	Storage    types.String `tfsdk:"storage"`
	SizeGB     types.Int64  `tfsdk:"size_gb"`
	Content    types.String `tfsdk:"content"`
	ImportFrom types.String `tfsdk:"import_from"`
	Readonly   types.Bool   `tfsdk:"readonly"`
}

type Network struct {
	ID       types.Int64  `tfsdk:"id"`
	Bridge   types.String `tfsdk:"bridge"`
	Firewall types.Bool   `tfsdk:"firewall"`
}

type resourceNodeVirtualMachineModel struct {
	ID       types.Int64    `tfsdk:"id"`
	Reboot   types.Bool     `tfsdk:"reboot"`
	FWConfig types.String   `tfsdk:"fw_config"`
	Node     types.String   `tfsdk:"node"`
	Ides     []*Disk        `tfsdk:"ide"`
	Scsis    []*Disk        `tfsdk:"scsi"`
	Networks []*Network     `tfsdk:"network"`
	Memory   types.Int64    `tfsdk:"memory"`
	CPUs     types.Int64    `tfsdk:"cpus"`
	Serials  []types.String `tfsdk:"serials"`
}

type resourceNodeVirtualMachine struct {
	t *tasks.Client
	q *qemu.Client
	c *status.Client
}

func (r *resourceNodeVirtualMachine) SetClient(p *proxmox.Client) {
	r.t = tasks.New(p)
	r.q = qemu.New(p)
	r.c = status.New(p)
}

func (r *resourceNodeVirtualMachine) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_virtual_machine"
}

func (e *resourceNodeVirtualMachine) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	diskBlock := func(format string, min int64, max int64) schema.ListNestedBlock {
		return schema.ListNestedBlock{
			Description: fmt.Sprintf("A %s disk object", format),
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Optional:    true,
						Description: fmt.Sprintf("The disk ID for this disk (i.e. %s0 would just be 0)", format),
						Validators: []validator.Int64{
							validators.NotEqual(path.MatchRelative().AtParent().AtName("id")),
							int64validator.Between(min, max),
						},
					},
					"volume_id": schema.StringAttribute{
						Computed:    true,
						Description: "The volume ID for this disk",
					},
					"content": schema.StringAttribute{
						Optional:    true,
						Description: "The content ID for this disk",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("storage")),
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("import_from")),
						},
					},
					"size_gb": schema.Int64Attribute{
						Optional:    true,
						Description: "The size in GB if creating a disk",
						Validators: []validator.Int64{
							int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("storage")),
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("import_from")),
						},
					},
					"storage": schema.StringAttribute{
						Optional:    true,
						Description: "The node storage ID to place the new disk",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("storage")),
							stringvalidator.Any(
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("import_from")),
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("size_gb")),
							),
						},
					},
					"import_from": schema.StringAttribute{
						Optional:    true,
						Description: "A volid of an existing disk to copy from",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("size_gb")),
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("content")),
						},
					},
					"readonly": schema.BoolAttribute{
						Optional:    true,
						Description: "If set will put the disk in 'snapshot' mode making it readonly",
					},
				},
			},
		}
	}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:    true,
				Description: "The vmid of the VM",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory": schema.Int64Attribute{
				Required:    true,
				Description: "Memory allocation in MB",
			},
			"cpus": schema.Int64Attribute{
				Required:    true,
				Description: "The number of cpus/cores to allocate",
			},
			"node": schema.StringAttribute{
				Required:    true,
				Description: "The name of the node to schedule the VM onto",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fw_config": schema.StringAttribute{
				Optional:    true,
				Description: "Additional arguments to pass to qemu",
			},
			"reboot": schema.BoolAttribute{
				Optional:    true,
				Description: "Reboot on config change",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"serials": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list (max 3) of serial devices on the guest",
			},
		},
		Blocks: map[string]schema.Block{
			"ide":  diskBlock("ide", 0, 3),
			"scsi": diskBlock("scsi", 0, 30),
			"network": schema.ListNestedBlock{
				Description: "A network interface",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Optional:    true,
							Description: "The network id (i.e. net0 would be 0)",
						},
						"bridge": schema.StringAttribute{
							Required:    true,
							Description: "The hosts network bridge to use",
						},
						"firewall": schema.BoolAttribute{
							Required:    true,
							Description: "If set will utilize the proxmox firewall",
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

	creq := qemu.CreateRequest{
		Node: plan.Node.ValueString(),
		Vmid: int(plan.ID.ValueInt64()),

		Memory: proxmox.Int(int(plan.Memory.ValueInt64())),
		Cores:  proxmox.Int(int(plan.CPUs.ValueInt64())),
	}

	serials := make([]*string, len(plan.Serials))
	for i, s := range plan.Serials {
		serialString := s.ValueString()
		serials[i] = &serialString
	}
	if len(serials) > 0 {
		creq.Serials = (*qemu.Serials)(&serials)
	}

	if plan.FWConfig.ValueString() != "" {
		cfgString := "-fw_cfg " + plan.FWConfig.ValueString()
		creq.Args = &cfgString
	}

	nets := make(qemu.Nets, 0)
	for _, net := range plan.Networks {
		for net.ID.ValueInt64() > int64(len(nets)-1) {
			nets = append(nets, nil)
		}
		n := &qemu.Net{
			Firewall: proxmox.PVEBool(net.Firewall.ValueBool()),
			Bridge:   proxmox.String(net.Bridge.ValueString()),
			Model:    qemu.NetModel_VIRTIO,
		}
		nets[net.ID.ValueInt64()] = n
	}
	creq.Nets = &nets

	if len(plan.Ides) > 0 {
		ideArr := make(qemu.Ides, 0)
		for _, d := range plan.Ides {
			for d.ID.ValueInt64() > int64(len(ideArr)-1) {
				ideArr = append(ideArr, nil)
			}
			ide := &qemu.Ide{}
			proxmoxDisk(d, (*wrappedIde)(ide))
			ideArr[d.ID.ValueInt64()] = ide
		}
		creq.Ides = &ideArr
	}

	if len(plan.Scsis) > 0 {
		scsiArr := make(qemu.Scsis, 0)
		for _, d := range plan.Scsis {
			for d.ID.ValueInt64() > int64(len(scsiArr)-1) {
				scsiArr = append(scsiArr, nil)
			}
			scsi := &qemu.Scsi{}
			proxmoxDisk(d, (*wrappedScsi)(scsi))
			scsiArr[d.ID.ValueInt64()] = scsi
		}
		creq.Scsis = &scsiArr
	}

	task, err := r.q.Create(ctx, creq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating VM",
			"An unexpected error occurred when creating the VM. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}

	diags = r.t.Wait(ctx, task, plan.Node.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.q.VmConfig(ctx, qemu.VmConfigRequest{
		Node: plan.Node.ValueString(),
		Vmid: int(plan.ID.ValueInt64()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error gettng  VM config",
			"An unexpected error occurred when retreiving the VM config. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}
	for _, d := range plan.Scsis {
		if config.Scsis == nil || len(*config.Scsis) <= int(d.ID.ValueInt64()) {
			resp.Diagnostics.AddError(
				"Not enough disks",
				"Something went wrong creating the VM not enough scsi Disks",
			)
			return
		}
		d.VolumeID = types.StringValue((*config.Scsis)[d.ID.ValueInt64()].File)
	}
	for _, d := range plan.Ides {
		if config.Ides == nil || len(*config.Ides) <= int(d.ID.ValueInt64()) {
			resp.Diagnostics.AddError(
				"Not enough disks",
				"Something went wrong creating the VM not enough ide Disks",
			)
			return
		}
		d.VolumeID = types.StringValue((*config.Ides)[d.ID.ValueInt64()].File)
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
	if err != nil {
		if err.Error() == fmt.Sprintf("non 200: 500 Configuration file 'nodes/%s/qemu-server/%d.conf' does not exist", data.Node.ValueString(), data.ID.ValueInt64()) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting VM",
			"An unexpected error occurred when deleting the VM. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}
	diags := r.t.Wait(ctx, taskID, data.Node.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
	if plan.FWConfig.ValueString() != "" {
		cfgString := "-fw_cfg " + plan.FWConfig.ValueString()
		configReq.Args = &cfgString
	}
	ideArr := make(qemu.Ides, len(plan.Ides))
	for i, d := range plan.Ides {
		ide := &qemu.Ide{}
		proxmoxDisk(d, (*wrappedIde)(ide))
		ideArr[i] = ide
	}

	scsiArr := make(qemu.Scsis, len(plan.Scsis))
	for i, d := range plan.Scsis {
		scsi := &qemu.Scsi{}
		proxmoxDisk(d, (*wrappedScsi)(scsi))
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

	diags = r.t.Wait(ctx, task, plan.Node.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Reboot.ValueBool() {
		task, err = r.c.VmReboot(ctx, status.VmRebootRequest{
			Node:    plan.Node.ValueString(),
			Vmid:    int(plan.ID.ValueInt64()),
			Timeout: proxmox.Int(300),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error rebooting VM",
				"An unexpected error occurred when rebooting the VM. "+
					"Proxmox API Error: "+err.Error(),
			)
			return
		}
		diags = r.t.Wait(ctx, task, plan.Node.ValueString())
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
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
		if err.Error() == fmt.Sprintf("non 200: 500 Configuration file 'nodes/%s/qemu-server/%d.conf' does not exist", state.Node.ValueString(), state.ID.ValueInt64()) {
			return
		}
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
	if state.Ides == nil {
		state.Ides = make([]*Disk, 0)
	}
	if config.Ides != nil {
		for i, ide := range *config.Ides {
			if len(state.Ides) <= i {
				state.Ides = append(state.Ides, &Disk{})
			}
			if ide == nil {
				state.Ides[i] = nil
				continue
			}
			diags := state.Ides[i].buildDisk(i, ide.File, (*bool)(ide.Snapshot))
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}
	if state.Scsis == nil {
		state.Scsis = make([]*Disk, 0)
	}
	if config.Scsis != nil {
		for i, scsi := range *config.Scsis {
			if len(state.Scsis) <= i {
				state.Scsis = append(state.Scsis, &Disk{})
			}
			diags := state.Scsis[i].buildDisk(i, scsi.File, (*bool)(scsi.Snapshot))
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	if config.Nets != nil {
		state.Networks = make([]*Network, len(*config.Nets))
		for i, net := range *config.Nets {
			n := &Network{}
			if net.Firewall != nil {
				n.Firewall = types.BoolValue(bool(*net.Firewall))
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

func (d *Disk) buildDisk(i int, file string, snapshot *bool) diag.Diagnostics {
	diags := diag.Diagnostics{}
	storageID := strings.Split(file, ":")[0]
	snapshotBool := false
	if snapshot != nil {
		snapshotBool = *snapshot
	}
	d.Storage = types.StringValue(storageID)
	d.Content = types.StringValue(file)
	d.Readonly = types.BoolValue(snapshotBool)
	return diags
}

func proxmoxDisk(d *Disk, qd wrappedDisk) {
	if d.Content.ValueString() != "" {
		qd.SetFile(d.Content.ValueString())
		if filepath.Ext(d.Content.ValueString()) == ".iso" {
			qd.SetMedia(string(qemu.IdeMedia_CDROM))
		}
	} else if d.SizeGB.ValueInt64() != 0 {
		qd.SetFile(fmt.Sprintf("%s:%d", d.Storage.ValueString(), d.SizeGB.ValueInt64()))
	} else if d.ImportFrom.ValueString() != "" {
		qd.SetFile(d.Storage.ValueString() + ":0")
		qd.SetImportFrom(d.ImportFrom.ValueString())
		qd.UnSetMedia()
	}

	qd.SetSnapshot(d.Readonly.ValueBool())
}
