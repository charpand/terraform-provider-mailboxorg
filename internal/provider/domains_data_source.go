package provider

import (
	"context"
	"fmt"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/domains"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &domainsDataSource{}

// NewDomainsDataSource returns the mailboxorg_domains data source.
func NewDomainsDataSource() datasource.DataSource {
	return &domainsDataSource{}
}

type domainsDataSource struct {
	client *client.Client
}

type domainsModel struct {
	ID      types.String `tfsdk:"id"`
	Account types.String `tfsdk:"account"`
	Filter  types.String `tfsdk:"filter"`
	Domains types.List   `tfsdk:"domains"`
}

var domainEntryType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"domain":      types.StringType,
	"count_mails": types.Int64Type,
}}

func (d *domainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *domainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the domains of a mailbox.org account (`domain.list`).",
		Attributes: map[string]schema.Attribute{
			"id":      schema.StringAttribute{Computed: true, Description: "The account name."},
			"account": schema.StringAttribute{Required: true, Description: "The account whose domains are listed."},
			"filter":  schema.StringAttribute{Optional: true, Description: "Search filter passed to the API."},
			"domains": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The domains of the account.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"domain":      schema.StringAttribute{Computed: true, Description: "The domain name."},
					"count_mails": schema.Int64Attribute{Computed: true, Description: "Number of mail addresses defined for the domain."},
				}},
			},
		},
	}
}

func (d *domainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData))
		return
	}
	d.client = c
}

func (d *domainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state domainsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := domains.List(ctx, d.client, state.Account.ValueString(), state.Filter.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to list domains", err.Error())
		return
	}

	values := make([]attr.Value, len(entries))
	for i, e := range entries {
		obj, diags := types.ObjectValue(domainEntryType.AttrTypes, map[string]attr.Value{
			"domain":      types.StringValue(e.Domain),
			"count_mails": types.Int64Value(int64(e.CountMails)),
		})
		resp.Diagnostics.Append(diags...)
		values[i] = obj
	}
	list, diags := types.ListValue(domainEntryType, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = state.Account
	state.Domains = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
