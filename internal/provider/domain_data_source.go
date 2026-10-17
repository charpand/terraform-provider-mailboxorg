package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/domains"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &domainDataSource{}

// NewDomainDataSource returns the mailboxorg_domain data source.
func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

type domainDataSource struct {
	client *client.Client
}

type domainModel struct {
	ID             types.String `tfsdk:"id"`
	Domain         types.String `tfsdk:"domain"`
	Status         types.String `tfsdk:"status"`
	ContextID      types.String `tfsdk:"context_id"`
	DateRegister   types.String `tfsdk:"date_register"`
	Memo           types.String `tfsdk:"memo"`
	PlansAvailable types.List   `tfsdk:"plans_available"`
}

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a domain known to mailbox.org (`domain.get`).",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, Description: "The domain name."},
			"domain":          schema.StringAttribute{Required: true, Description: "The domain name to look up."},
			"status":          schema.StringAttribute{Computed: true, Description: "Domain status as reported by the API."},
			"context_id":      schema.StringAttribute{Computed: true, Description: "Context (cloud office) ID the domain belongs to."},
			"date_register":   schema.StringAttribute{Computed: true, Description: "Registration date as reported by the API."},
			"memo":            schema.StringAttribute{Computed: true, Description: "Free-text memo."},
			"plans_available": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Mailbox plans available for this domain."},
		},
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := domains.Get(ctx, d.client, state.Domain.ValueString())
	if errors.Is(err, client.ErrForbidden) {
		resp.Diagnostics.AddAttributeError(path.Root("domain"), "Domain not found or not manageable",
			fmt.Sprintf("mailbox.org refused access to %q (%s). The API gives this answer both for domains that do not exist and for domains this account may not manage.", state.Domain.ValueString(), err))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read domain", err.Error())
		return
	}

	plans := make([]string, len(res.PlansAvailable))
	for i, p := range res.PlansAvailable {
		plans[i] = string(p)
	}
	planList, diags := types.ListValueFrom(ctx, types.StringType, plans)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = state.Domain
	state.Status = types.StringValue(string(res.Status))
	state.ContextID = types.StringValue(string(res.ContextID))
	state.DateRegister = types.StringValue(string(res.DateRegister))
	state.Memo = types.StringValue(string(res.Memo))
	state.PlansAvailable = planList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
