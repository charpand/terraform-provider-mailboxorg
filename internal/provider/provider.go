// Package provider implements the mailbox.org Terraform provider.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &mailboxorgProvider{}

type mailboxorgProvider struct {
	version string
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
	resp.Schema = schema.Schema{}
}

func (p *mailboxorgProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *mailboxorgProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

func (p *mailboxorgProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
