// Package provider implements the mailbox.org Terraform provider.
package provider

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &mailboxorgProvider{}

type mailboxorgProvider struct {
	version string
}

type providerModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	BaseURL  types.String `tfsdk:"base_url"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

// New returns a constructor for the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mailboxorgProvider{version: version}
	}
}

func (p *mailboxorgProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mailboxorg"
	resp.Version = p.version
}

func (p *mailboxorgProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage mailbox.org through its JSON-RPC API.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "mailbox.org API user. Can also be set with the MAILBOXORG_USERNAME environment variable.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password of the API user. Can also be set with the MAILBOXORG_PASSWORD environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "API endpoint. Defaults to " + client.DefaultBaseURL + ". Can also be set with MAILBOXORG_BASE_URL.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "HTTP request timeout in seconds. Defaults to 60. Can also be set with MAILBOXORG_TIMEOUT.",
			},
		},
	}
}

func (p *mailboxorgProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for name, v := range map[string]interface{ IsUnknown() bool }{
		"username": cfg.Username, "password": cfg.Password, "base_url": cfg.BaseURL, "timeout": cfg.Timeout,
	} {
		if v.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root(name), "Unknown provider value",
				"The value is not known yet. Set it to a static value or make it known before the provider is configured.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	username := stringOrEnv(cfg.Username, "MAILBOXORG_USERNAME")
	password := stringOrEnv(cfg.Password, "MAILBOXORG_PASSWORD")
	baseURL := stringOrEnv(cfg.BaseURL, "MAILBOXORG_BASE_URL")

	timeout := time.Duration(0)
	if !cfg.Timeout.IsNull() {
		timeout = time.Duration(cfg.Timeout.ValueInt64()) * time.Second
	} else if env := os.Getenv("MAILBOXORG_TIMEOUT"); env != "" {
		secs, err := strconv.Atoi(env)
		if err != nil || secs <= 0 {
			resp.Diagnostics.AddError("Invalid MAILBOXORG_TIMEOUT", "MAILBOXORG_TIMEOUT must be a positive number of seconds.")
			return
		}
		timeout = time.Duration(secs) * time.Second
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(path.Root("username"), "Missing username",
			"Set the username attribute or the MAILBOXORG_USERNAME environment variable.")
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Missing password",
			"Set the password attribute or the MAILBOXORG_PASSWORD environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// No network call here: the client logs in on first use.
	c, err := client.New(client.Config{BaseURL: baseURL, Username: username, Password: password, Timeout: timeout})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create mailbox.org client", err.Error())
		return
	}
	resp.DataSourceData = c
	resp.ResourceData = c
}

func stringOrEnv(v types.String, env string) string {
	if !v.IsNull() {
		return v.ValueString()
	}
	return os.Getenv(env)
}

func (p *mailboxorgProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *mailboxorgProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewDomainDataSource}
}
