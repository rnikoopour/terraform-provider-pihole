package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

type PiholeProvider struct {
	version string
}

type PiholeProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *PiholeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pihole"
	resp.Version = p.version
}

func (p *PiholeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Pi-hole v6 resources.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "Base URL of the Pi-hole server (e.g. https://192.0.2.1).",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Pi-hole web password or app password (if 2FA is enabled).",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS certificate verification. Useful for self-signed certificates.",
			},
		},
	}
}

func (p *PiholeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config PiholeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	insecure := !config.Insecure.IsNull() && config.Insecure.ValueBool()
	client := pihole.NewClient(config.URL.ValueString(), config.Password.ValueString(), insecure)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PiholeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewListResource,
		NewDomainResource,
		NewGroupResource,
		NewDNSRecordResource,
		NewCNAMERecordResource,
		NewConfigResource,
		NewGravityResource,
	}
}

func (p *PiholeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PiholeProvider{version: version}
	}
}
