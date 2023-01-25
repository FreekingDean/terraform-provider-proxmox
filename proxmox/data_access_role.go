package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/roles"
)

var (
	_ datasource.DataSource = &accessRoleDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessRoleDataSource)
}

type accessRoleDataSource struct {
	client *roles.Client
}

func NewAccessRoleDataSource() datasource.DataSource {
	return &accessRoleDataSource{}
}

type accessRoleModel struct {
	Roleid                    types.String `tfsdk:"roleid"`
	DatastoreAllocate         types.Bool   `tfsdk:"datastoreallocate"`
	DatastoreAllocatespace    types.Bool   `tfsdk:"datastoreallocatespace"`
	DatastoreAllocatetemplate types.Bool   `tfsdk:"datastoreallocatetemplate"`
	DatastoreAudit            types.Bool   `tfsdk:"datastoreaudit"`
	GroupAllocate             types.Bool   `tfsdk:"groupallocate"`
	PermissionsModify         types.Bool   `tfsdk:"permissionsmodify"`
	PoolAllocate              types.Bool   `tfsdk:"poolallocate"`
	PoolAudit                 types.Bool   `tfsdk:"poolaudit"`
	RealmAllocate             types.Bool   `tfsdk:"realmallocate"`
	RealmAllocateuser         types.Bool   `tfsdk:"realmallocateuser"`
	SdnAllocate               types.Bool   `tfsdk:"sdnallocate"`
	SdnAudit                  types.Bool   `tfsdk:"sdnaudit"`
	SysAudit                  types.Bool   `tfsdk:"sysaudit"`
	SysConsole                types.Bool   `tfsdk:"sysconsole"`
	SysIncoming               types.Bool   `tfsdk:"sysincoming"`
	SysModify                 types.Bool   `tfsdk:"sysmodify"`
	SysPowermgmt              types.Bool   `tfsdk:"syspowermgmt"`
	SysSyslog                 types.Bool   `tfsdk:"syssyslog"`
	UserModify                types.Bool   `tfsdk:"usermodify"`
	VmAllocate                types.Bool   `tfsdk:"vmallocate"`
	VmAudit                   types.Bool   `tfsdk:"vmaudit"`
	VmBackup                  types.Bool   `tfsdk:"vmbackup"`
	VmClone                   types.Bool   `tfsdk:"vmclone"`
	VmConfigCdrom             types.Bool   `tfsdk:"vmconfigcdrom"`
	VmConfigCloudinit         types.Bool   `tfsdk:"vmconfigcloudinit"`
	VmConfigCpu               types.Bool   `tfsdk:"vmconfigcpu"`
	VmConfigDisk              types.Bool   `tfsdk:"vmconfigdisk"`
	VmConfigHwtype            types.Bool   `tfsdk:"vmconfighwtype"`
	VmConfigMemory            types.Bool   `tfsdk:"vmconfigmemory"`
	VmConfigNetwork           types.Bool   `tfsdk:"vmconfignetwork"`
	VmConfigOptions           types.Bool   `tfsdk:"vmconfigoptions"`
	VmConsole                 types.Bool   `tfsdk:"vmconsole"`
	VmMigrate                 types.Bool   `tfsdk:"vmmigrate"`
	VmMonitor                 types.Bool   `tfsdk:"vmmonitor"`
	VmPowermgmt               types.Bool   `tfsdk:"vmpowermgmt"`
	VmSnapshot                types.Bool   `tfsdk:"vmsnapshot"`
	VmSnapshotRollback        types.Bool   `tfsdk:"vmsnapshotrollback"`
}

func (d *accessRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = roles.New(client)
	}
}

func (d *accessRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_role"
}

func (d *accessRoleDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"roleid": schema.StringAttribute{
				Required: true,
			},
			"datastoreallocate": schema.BoolAttribute{
				Computed: true,
			},
			"datastoreallocatespace": schema.BoolAttribute{
				Computed: true,
			},
			"datastoreallocatetemplate": schema.BoolAttribute{
				Computed: true,
			},
			"datastoreaudit": schema.BoolAttribute{
				Computed: true,
			},
			"groupallocate": schema.BoolAttribute{
				Computed: true,
			},
			"permissionsmodify": schema.BoolAttribute{
				Computed: true,
			},
			"poolallocate": schema.BoolAttribute{
				Computed: true,
			},
			"poolaudit": schema.BoolAttribute{
				Computed: true,
			},
			"realmallocate": schema.BoolAttribute{
				Computed: true,
			},
			"realmallocateuser": schema.BoolAttribute{
				Computed: true,
			},
			"sdnallocate": schema.BoolAttribute{
				Computed: true,
			},
			"sdnaudit": schema.BoolAttribute{
				Computed: true,
			},
			"sysaudit": schema.BoolAttribute{
				Computed: true,
			},
			"sysconsole": schema.BoolAttribute{
				Computed: true,
			},
			"sysincoming": schema.BoolAttribute{
				Computed: true,
			},
			"sysmodify": schema.BoolAttribute{
				Computed: true,
			},
			"syspowermgmt": schema.BoolAttribute{
				Computed: true,
			},
			"syssyslog": schema.BoolAttribute{
				Computed: true,
			},
			"usermodify": schema.BoolAttribute{
				Computed: true,
			},
			"vmallocate": schema.BoolAttribute{
				Computed: true,
			},
			"vmaudit": schema.BoolAttribute{
				Computed: true,
			},
			"vmbackup": schema.BoolAttribute{
				Computed: true,
			},
			"vmclone": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigcdrom": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigcloudinit": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigcpu": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigdisk": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfighwtype": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigmemory": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfignetwork": schema.BoolAttribute{
				Computed: true,
			},
			"vmconfigoptions": schema.BoolAttribute{
				Computed: true,
			},
			"vmconsole": schema.BoolAttribute{
				Computed: true,
			},
			"vmmigrate": schema.BoolAttribute{
				Computed: true,
			},
			"vmmonitor": schema.BoolAttribute{
				Computed: true,
			},
			"vmpowermgmt": schema.BoolAttribute{
				Computed: true,
			},
			"vmsnapshot": schema.BoolAttribute{
				Computed: true,
			},
			"vmsnapshotrollback": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *accessRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessRoleModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessRole, err := d.client.Find(
		ctx,
		&roles.FindRequest{
			Roleid: state.Roleid.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessRole",
			err.Error(),
		)
		return
	}

	if accessRole.DatastoreAllocate != nil {
		state.DatastoreAllocate = types.BoolValue(bool(*accessRole.DatastoreAllocate))
	}
	if accessRole.DatastoreAllocatespace != nil {
		state.DatastoreAllocatespace = types.BoolValue(bool(*accessRole.DatastoreAllocatespace))
	}
	if accessRole.DatastoreAllocatetemplate != nil {
		state.DatastoreAllocatetemplate = types.BoolValue(bool(*accessRole.DatastoreAllocatetemplate))
	}
	if accessRole.DatastoreAudit != nil {
		state.DatastoreAudit = types.BoolValue(bool(*accessRole.DatastoreAudit))
	}
	if accessRole.GroupAllocate != nil {
		state.GroupAllocate = types.BoolValue(bool(*accessRole.GroupAllocate))
	}
	if accessRole.PermissionsModify != nil {
		state.PermissionsModify = types.BoolValue(bool(*accessRole.PermissionsModify))
	}
	if accessRole.PoolAllocate != nil {
		state.PoolAllocate = types.BoolValue(bool(*accessRole.PoolAllocate))
	}
	if accessRole.PoolAudit != nil {
		state.PoolAudit = types.BoolValue(bool(*accessRole.PoolAudit))
	}
	if accessRole.RealmAllocate != nil {
		state.RealmAllocate = types.BoolValue(bool(*accessRole.RealmAllocate))
	}
	if accessRole.RealmAllocateuser != nil {
		state.RealmAllocateuser = types.BoolValue(bool(*accessRole.RealmAllocateuser))
	}
	if accessRole.SdnAllocate != nil {
		state.SdnAllocate = types.BoolValue(bool(*accessRole.SdnAllocate))
	}
	if accessRole.SdnAudit != nil {
		state.SdnAudit = types.BoolValue(bool(*accessRole.SdnAudit))
	}
	if accessRole.SysAudit != nil {
		state.SysAudit = types.BoolValue(bool(*accessRole.SysAudit))
	}
	if accessRole.SysConsole != nil {
		state.SysConsole = types.BoolValue(bool(*accessRole.SysConsole))
	}
	if accessRole.SysIncoming != nil {
		state.SysIncoming = types.BoolValue(bool(*accessRole.SysIncoming))
	}
	if accessRole.SysModify != nil {
		state.SysModify = types.BoolValue(bool(*accessRole.SysModify))
	}
	if accessRole.SysPowermgmt != nil {
		state.SysPowermgmt = types.BoolValue(bool(*accessRole.SysPowermgmt))
	}
	if accessRole.SysSyslog != nil {
		state.SysSyslog = types.BoolValue(bool(*accessRole.SysSyslog))
	}
	if accessRole.UserModify != nil {
		state.UserModify = types.BoolValue(bool(*accessRole.UserModify))
	}
	if accessRole.VmAllocate != nil {
		state.VmAllocate = types.BoolValue(bool(*accessRole.VmAllocate))
	}
	if accessRole.VmAudit != nil {
		state.VmAudit = types.BoolValue(bool(*accessRole.VmAudit))
	}
	if accessRole.VmBackup != nil {
		state.VmBackup = types.BoolValue(bool(*accessRole.VmBackup))
	}
	if accessRole.VmClone != nil {
		state.VmClone = types.BoolValue(bool(*accessRole.VmClone))
	}
	if accessRole.VmConfigCdrom != nil {
		state.VmConfigCdrom = types.BoolValue(bool(*accessRole.VmConfigCdrom))
	}
	if accessRole.VmConfigCloudinit != nil {
		state.VmConfigCloudinit = types.BoolValue(bool(*accessRole.VmConfigCloudinit))
	}
	if accessRole.VmConfigCpu != nil {
		state.VmConfigCpu = types.BoolValue(bool(*accessRole.VmConfigCpu))
	}
	if accessRole.VmConfigDisk != nil {
		state.VmConfigDisk = types.BoolValue(bool(*accessRole.VmConfigDisk))
	}
	if accessRole.VmConfigHwtype != nil {
		state.VmConfigHwtype = types.BoolValue(bool(*accessRole.VmConfigHwtype))
	}
	if accessRole.VmConfigMemory != nil {
		state.VmConfigMemory = types.BoolValue(bool(*accessRole.VmConfigMemory))
	}
	if accessRole.VmConfigNetwork != nil {
		state.VmConfigNetwork = types.BoolValue(bool(*accessRole.VmConfigNetwork))
	}
	if accessRole.VmConfigOptions != nil {
		state.VmConfigOptions = types.BoolValue(bool(*accessRole.VmConfigOptions))
	}
	if accessRole.VmConsole != nil {
		state.VmConsole = types.BoolValue(bool(*accessRole.VmConsole))
	}
	if accessRole.VmMigrate != nil {
		state.VmMigrate = types.BoolValue(bool(*accessRole.VmMigrate))
	}
	if accessRole.VmMonitor != nil {
		state.VmMonitor = types.BoolValue(bool(*accessRole.VmMonitor))
	}
	if accessRole.VmPowermgmt != nil {
		state.VmPowermgmt = types.BoolValue(bool(*accessRole.VmPowermgmt))
	}
	if accessRole.VmSnapshot != nil {
		state.VmSnapshot = types.BoolValue(bool(*accessRole.VmSnapshot))
	}
	if accessRole.VmSnapshotRollback != nil {
		state.VmSnapshotRollback = types.BoolValue(bool(*accessRole.VmSnapshotRollback))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
