package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/mails"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &mailDataSource{}

// codeDatabaseProblem is the generic "database problem" error; mail.get uses it
// for an unknown address (verified live), other methods use it for other failures.
const codeDatabaseProblem = 6

// NewMailDataSource returns the mailboxorg_mail data source.
func NewMailDataSource() datasource.DataSource {
	return &mailDataSource{}
}

type mailDataSource struct {
	client *client.Client
}

type mailModel struct {
	ID                   types.String `tfsdk:"id"`
	Mail                 types.String `tfsdk:"mail"`
	UID                  types.String `tfsdk:"uid"`
	Type                 types.String `tfsdk:"type"`
	Plan                 types.String `tfsdk:"plan"`
	PlansAvailable       types.List   `tfsdk:"plans_available"`
	AdditionalMailQuota  types.String `tfsdk:"additional_mail_quota"`
	AdditionalCloudQuota types.String `tfsdk:"additional_cloud_quota"`
	FirstName            types.String `tfsdk:"first_name"`
	LastName             types.String `tfsdk:"last_name"`
	InboxSave            types.Bool   `tfsdk:"inboxsave"`
	RequireResetPassword types.Bool   `tfsdk:"require_reset_password"`
	Forwards             types.List   `tfsdk:"forwards"`
	Aliases              types.List   `tfsdk:"aliases"`
	AlternateMail        types.String `tfsdk:"alternate_mail"`
	Memo                 types.String `tfsdk:"memo"`
	AllowNets            types.String `tfsdk:"allow_nets"`
	UIDExtern            types.String `tfsdk:"uid_extern"`
	Active               types.Bool   `tfsdk:"active"`
	Capabilities         types.List   `tfsdk:"capabilities"`
	Language             types.String `tfsdk:"language"`
	CreationDate         types.String `tfsdk:"creation_date"`
	DeletionDate         types.String `tfsdk:"deletion_date"`
	MailArchiveEnabled   types.Bool   `tfsdk:"mailarchive_enabled"`
}

func (d *mailDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mail"
}

func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Computed: true, Description: desc}
}

func computedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{Computed: true, Description: desc}
}

func computedStrings(desc string) schema.ListAttribute {
	return schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: desc}
}

func (d *mailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a mail address (mailbox, forward or catch-all) at mailbox.org (`mail.get`). " +
			"Password and postal address data are never exposed.",
		Attributes: map[string]schema.Attribute{
			"id":                     computedString("The mail address."),
			"mail":                   schema.StringAttribute{Required: true, Description: "The mail address to look up."},
			"uid":                    computedString("Internal user ID."),
			"type":                   computedString("Address type: `forward`, `inbox` or `inboxforward`."),
			"plan":                   computedString("Current plan: `light`, `standard` or `premium`."),
			"plans_available":        computedStrings("Plans this address can be switched to."),
			"additional_mail_quota":  computedString("Additional mail quota."),
			"additional_cloud_quota": computedString("Additional cloud quota."),
			"first_name":             computedString("First name."),
			"last_name":              computedString("Last name."),
			"inboxsave":              computedBool("Whether forwarded mail is also kept in the inbox."),
			"require_reset_password": computedBool("Whether the user must reset the password at next login."),
			"forwards":               computedStrings("Forwarding targets."),
			"aliases":                computedStrings("Alias addresses."),
			"alternate_mail":         computedString("Alternate (recovery) mail address."),
			"memo":                   computedString("Free-text memo."),
			"allow_nets":             computedString("Networks allowed to log in."),
			"uid_extern":             computedString("External user ID."),
			"active":                 computedBool("Whether the address is active."),
			"capabilities":           computedStrings("Enabled capabilities, such as `MAIL_SPAMPROTECTION`."),
			"language":               computedString("Language (RFC 5646)."),
			"creation_date":          computedString("Creation date (YYYY-MM-DD)."),
			"deletion_date":          computedString("Planned deletion date, if any."),
			"mailarchive_enabled":    computedBool("Whether the mail archive is enabled."),
		},
	}
}

func (d *mailDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state mailModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := mails.Get(ctx, d.client, state.Mail.ValueString())
	var apiErr *client.APIError
	if errors.Is(err, client.ErrForbidden) || (errors.As(err, &apiErr) && apiErr.Code == codeDatabaseProblem) {
		resp.Diagnostics.AddAttributeError(path.Root("mail"), "Mail address not found or not manageable",
			fmt.Sprintf("mailbox.org could not provide %q (%s). The API answers like this both for addresses that do not exist (error 6) and for addresses this login may not manage (error 8).", state.Mail.ValueString(), err))
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read mail address", err.Error())
		return
	}

	list := func(l jsonx.TextList) types.List {
		v, diags := types.ListValueFrom(ctx, types.StringType, l.Strings())
		resp.Diagnostics.Append(diags...)
		return v
	}
	text := func(t jsonx.Text) types.String { return types.StringValue(string(t)) }

	state.ID = state.Mail
	state.UID = text(m.UID)
	state.Type = text(m.Type)
	state.Plan = text(m.Plan)
	state.PlansAvailable = list(m.PlansAvailable)
	state.AdditionalMailQuota = text(m.AdditionalMailQuota)
	state.AdditionalCloudQuota = text(m.AdditionalCloudQuota)
	state.FirstName = text(m.FirstName)
	state.LastName = text(m.LastName)
	state.InboxSave = types.BoolValue(bool(m.InboxSave))
	state.RequireResetPassword = types.BoolValue(bool(m.RequireResetPassword))
	state.Forwards = list(m.Forwards)
	state.Aliases = list(m.Aliases)
	state.AlternateMail = text(m.AlternateMail)
	state.Memo = text(m.Memo)
	state.AllowNets = text(m.AllowNets)
	state.UIDExtern = text(m.UIDExtern)
	state.Active = types.BoolValue(bool(m.Active))
	state.Capabilities = list(m.Capabilities)
	state.Language = text(m.Language)
	state.CreationDate = text(m.CreationDateISO)
	state.DeletionDate = text(m.DeletionDate)
	state.MailArchiveEnabled = types.BoolValue(bool(m.MailArchiveEnabled))
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
