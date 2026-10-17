package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/mails"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &mailsDataSource{}

// NewMailsDataSource returns the mailboxorg_mails data source.
func NewMailsDataSource() datasource.DataSource {
	return &mailsDataSource{}
}

type mailsDataSource struct {
	client *client.Client
}

type mailsModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Query  types.String `tfsdk:"query"`
	Mails  types.List   `tfsdk:"mails"`
}

var mailEntryType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"mail":          types.StringType,
	"domain":        types.StringType,
	"type":          types.StringType,
	"plan":          types.StringType,
	"memo":          types.StringType,
	"forwards":      types.ListType{ElemType: types.StringType},
	"aliases":       types.ListType{ElemType: types.StringType},
	"capabilities":  types.ListType{ElemType: types.StringType},
	"creation_date": types.StringType,
}}

func (d *mailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mails"
}

func (d *mailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists mail addresses at mailbox.org (`mail.list`), reading every page. " +
			"Set `domain`, `query`, or both.",
		Attributes: map[string]schema.Attribute{
			"id":     computedString("The domain or query that was listed."),
			"domain": schema.StringAttribute{Optional: true, Description: "List the addresses of this domain."},
			"query": schema.StringAttribute{
				Optional:    true,
				Description: "Search across all domains in addresses and names; `%` is a wildcard.",
			},
			"mails": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The matching addresses.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"mail":          computedString("The mail address."),
					"domain":        computedString("Its domain."),
					"type":          computedString("Address type: `forward`, `inbox` or `inboxforward`."),
					"plan":          computedString("Current plan."),
					"memo":          computedString("Free-text memo."),
					"forwards":      computedStrings("Forwarding targets."),
					"aliases":       computedStrings("Alias addresses."),
					"capabilities":  computedStrings("Enabled capabilities."),
					"creation_date": computedString("Creation date."),
				}},
			},
		},
	}
}

func (d *mailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state mailsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	domain, query := state.Domain.ValueString(), state.Query.ValueString()
	if strings.TrimSpace(domain) == "" && strings.TrimSpace(query) == "" {
		resp.Diagnostics.AddError("Missing selection", "Set at least one of `domain` or `query`.")
		return
	}

	entries, err := mails.List(ctx, d.client, mails.ListOptions{Domain: domain, Query: query})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list mail addresses", err.Error())
		return
	}

	strs := func(l jsonx.TextList) attr.Value {
		v, diags := types.ListValueFrom(ctx, types.StringType, l.Strings())
		resp.Diagnostics.Append(diags...)
		return v
	}
	values := make([]attr.Value, len(entries))
	for i, e := range entries {
		obj, diags := types.ObjectValue(mailEntryType.AttrTypes, map[string]attr.Value{
			"mail":          types.StringValue(e.Mail),
			"domain":        types.StringValue(string(e.Domain)),
			"type":          types.StringValue(string(e.Type)),
			"plan":          types.StringValue(string(e.Plan)),
			"memo":          types.StringValue(string(e.Memo)),
			"forwards":      strs(e.Forwards),
			"aliases":       strs(e.Aliases),
			"capabilities":  strs(e.Capabilities),
			"creation_date": types.StringValue(string(e.CreationDate)),
		})
		resp.Diagnostics.Append(diags...)
		values[i] = obj
	}
	list, diags := types.ListValue(mailEntryType, values)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(domain + "|" + query)
	state.Mails = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
