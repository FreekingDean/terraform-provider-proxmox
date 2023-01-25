package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/users"
)

var (
	_ datasource.DataSource = &accessUsersDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessUsersDataSource)
}

type accessUsersDataSource struct {
	client *users.Client
}

func NewAccessUsersDataSource() datasource.DataSource {
	return &accessUsersDataSource{}
}

type accessUsersModel struct {
	Enabled     types.Bool    `tfsdk:"enabled"`
	Full        types.Bool    `tfsdk:"full"`
	AccessUsers []AccessUsers `tfsdk:"access_users"`
}

type AccessUsers struct {
	Userid    types.String `tfsdk:"userid"`
	Comment   types.String `tfsdk:"comment"`
	Email     types.String `tfsdk:"email"`
	Enable    types.Bool   `tfsdk:"enable"`
	Expire    types.Int64  `tfsdk:"expire"`
	Firstname types.String `tfsdk:"firstname"`
	Groups    types.String `tfsdk:"groups"`
	Keys      types.String `tfsdk:"keys"`
	Lastname  types.String `tfsdk:"lastname"`
	RealmType types.String `tfsdk:"realmtype"`
	Tokens    []Tokens     `tfsdk:"tokens"`
}

type Tokens struct {
	Tokenid types.String `tfsdk:"tokenid"`
	Comment types.String `tfsdk:"comment"`
	Expire  types.Int64  `tfsdk:"expire"`
	Privsep types.Bool   `tfsdk:"privsep"`
}

func (d *accessUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = users.New(client)
	}
}

func (d *accessUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_users"
}

func (d *accessUsersDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"full": schema.BoolAttribute{
				Required: true,
			},
			"access_users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"userid": schema.StringAttribute{
							Computed: true,
						},
						"comment": schema.StringAttribute{
							Computed: true,
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"enable": schema.BoolAttribute{
							Computed: true,
						},
						"expire": schema.Int64Attribute{
							Computed: true,
						},
						"firstname": schema.StringAttribute{
							Computed: true,
						},
						"groups": schema.StringAttribute{
							Computed: true,
						},
						"keys": schema.StringAttribute{
							Computed: true,
						},
						"lastname": schema.StringAttribute{
							Computed: true,
						},
						"realmtype": schema.StringAttribute{
							Computed: true,
						},
						"tokens": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"tokenid": schema.StringAttribute{
										Computed: true,
									},
									"comment": schema.StringAttribute{
										Computed: true,
									},
									"expire": schema.Int64Attribute{
										Computed: true,
									},
									"privsep": schema.BoolAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *accessUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessUsersModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessUsers, err := d.client.Index(
		ctx,
		&users.IndexRequest{
			Enabled: state.Enabled.ValueBool(),
			Full:    state.Full.ValueBool(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessUsers",
			err.Error(),
		)
		return
	}

	for _, e := range *accessUsers {
		eState := AccessUsers{}
		eState.Userid = types.StringValue(e.Userid)
		if e.Comment != nil {
			eState.Comment = types.StringValue(*e.Comment)
		}
		if e.Email != nil {
			eState.Email = types.StringValue(*e.Email)
		}
		if e.Enable != nil {
			eState.Enable = types.BoolValue(bool(*e.Enable))
		}
		if e.Expire != nil {
			eState.Expire = types.Int64Value(*e.Expire)
		}
		if e.Firstname != nil {
			eState.Firstname = types.StringValue(*e.Firstname)
		}
		if e.Groups != nil {
			eState.Groups = types.StringValue(*e.Groups)
		}
		if e.Keys != nil {
			eState.Keys = types.StringValue(*e.Keys)
		}
		if e.Lastname != nil {
			eState.Lastname = types.StringValue(*e.Lastname)
		}
		if e.RealmType != nil {
			eState.RealmType = types.StringValue(*e.RealmType)
		}
		eState.Tokens = []TokensValue(e.Tokens)
		state.AccessUsers = append(state.AccessUsers, eState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
