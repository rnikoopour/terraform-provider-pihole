package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

// attrType maps — defined once, reused in schema defaults, objectdefault, and Read.
var (
	blockingAttrTypes = map[string]attr.Type{
		"active": types.BoolType,
		"mode":   types.StringType,
		"edns":   types.StringType,
	}
	rateLimitAttrTypes = map[string]attr.Type{
		"count":    types.Int64Type,
		"interval": types.Int64Type,
	}
	cacheAttrTypes = map[string]attr.Type{
		"size":      types.Int64Type,
		"optimizer": types.Int64Type,
	}
	specialDomainsAttrTypes = map[string]attr.Type{
		"mozilla_canary":       types.BoolType,
		"icloud_private_relay": types.BoolType,
		"designated_resolver":  types.BoolType,
	}
	interfaceAttrTypes = map[string]attr.Type{
		"theme": types.StringType,
		"boxed": types.BoolType,
	}
	dnsAttrTypes = map[string]attr.Type{
		"upstreams":          types.ListType{ElemType: types.StringType},
		"domain_needed":      types.BoolType,
		"expand_hosts":       types.BoolType,
		"bogus_priv":         types.BoolType,
		"dnssec":             types.BoolType,
		"query_logging":      types.BoolType,
		"cname_deep_inspect": types.BoolType,
		"block_esni":         types.BoolType,
		"host_record":        types.StringType,
		"listening_mode":     types.StringType,
		"interface":          types.StringType,
		"reply_when_busy":    types.StringType,
		"block_ttl":          types.Int64Type,
		"blocking":           types.ObjectType{AttrTypes: blockingAttrTypes},
		"rate_limit":         types.ObjectType{AttrTypes: rateLimitAttrTypes},
		"cache":              types.ObjectType{AttrTypes: cacheAttrTypes},
		"special_domains":    types.ObjectType{AttrTypes: specialDomainsAttrTypes},
	}
	webserverAttrTypes = map[string]attr.Type{
		"domain":    types.StringType,
		"interface": types.ObjectType{AttrTypes: interfaceAttrTypes},
	}
	miscAttrTypes = map[string]attr.Type{
		"privacy_level": types.Int64Type,
	}
	databaseAttrTypes = map[string]attr.Type{
		"max_db_days": types.Int64Type,
	}
)

// helper structs for unpacking/packing nested objects via .As() and ObjectValueFrom.
type dnsModel struct {
	Upstreams        types.List   `tfsdk:"upstreams"`
	DomainNeeded     types.Bool   `tfsdk:"domain_needed"`
	ExpandHosts      types.Bool   `tfsdk:"expand_hosts"`
	BogusPriv        types.Bool   `tfsdk:"bogus_priv"`
	DNSSEC           types.Bool   `tfsdk:"dnssec"`
	QueryLogging     types.Bool   `tfsdk:"query_logging"`
	CNAMEDeepInspect types.Bool   `tfsdk:"cname_deep_inspect"`
	BlockESNI        types.Bool   `tfsdk:"block_esni"`
	HostRecord       types.String `tfsdk:"host_record"`
	ListeningMode    types.String `tfsdk:"listening_mode"`
	Interface        types.String `tfsdk:"interface"`
	ReplyWhenBusy    types.String `tfsdk:"reply_when_busy"`
	BlockTTL         types.Int64  `tfsdk:"block_ttl"`
	Blocking         types.Object `tfsdk:"blocking"`
	RateLimit        types.Object `tfsdk:"rate_limit"`
	Cache            types.Object `tfsdk:"cache"`
	SpecialDomains   types.Object `tfsdk:"special_domains"`
}

type blockingModel struct {
	Active types.Bool   `tfsdk:"active"`
	Mode   types.String `tfsdk:"mode"`
	EDNS   types.String `tfsdk:"edns"`
}

type rateLimitModel struct {
	Count    types.Int64 `tfsdk:"count"`
	Interval types.Int64 `tfsdk:"interval"`
}

type cacheModel struct {
	Size      types.Int64 `tfsdk:"size"`
	Optimizer types.Int64 `tfsdk:"optimizer"`
}

type specialDomainsModel struct {
	MozillaCanary      types.Bool `tfsdk:"mozilla_canary"`
	ICloudPrivateRelay types.Bool `tfsdk:"icloud_private_relay"`
	DesignatedResolver types.Bool `tfsdk:"designated_resolver"`
}

type webserverModel struct {
	Domain    types.String `tfsdk:"domain"`
	Interface types.Object `tfsdk:"interface"`
}

type interfaceModel struct {
	Theme types.String `tfsdk:"theme"`
	Boxed types.Bool   `tfsdk:"boxed"`
}

type miscModel struct {
	PrivacyLevel types.Int64 `tfsdk:"privacy_level"`
}

type databaseModel struct {
	MaxDBDays types.Int64 `tfsdk:"max_db_days"`
}

type ConfigResource struct {
	client *pihole.Client
}

type ConfigResourceModel struct {
	ID       types.String `tfsdk:"id"`
	DNS      types.Object `tfsdk:"dns"`
	Webserver types.Object `tfsdk:"webserver"`
	Misc     types.Object `tfsdk:"misc"`
	Database types.Object `tfsdk:"database"`
}

func NewConfigResource() resource.Resource { return &ConfigResource{} }

func (r *ConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *ConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Pi-hole configuration settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"upstreams": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "Upstream DNS server addresses.",
					},
					"domain_needed": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Do not forward incomplete domain names to upstream DNS.",
					},
					"expand_hosts": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Add the local domain to simple hostnames in /etc/hosts.",
					},
					"bogus_priv": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Do not forward reverse lookups for private IP ranges to upstream.",
					},
					"dnssec": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable DNSSEC validation.",
					},
					"query_logging": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Log DNS queries.",
					},
					"cname_deep_inspect": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Follow CNAME chains when checking blocklists.",
					},
					"block_esni": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Block ESNI (Encrypted Server Name Indication).",
					},
					"host_record": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: `Custom host record for this Pi-hole instance, e.g. "pihole.example,192.0.2.1".`,
					},
					"listening_mode": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("LOCAL"),
						Description: `Interface listening mode. Valid values: "LOCAL" (allow only local requests), "SINGLE" (respond only on specified interface), "BIND" (bind only to specified interface), "ALL" (permit all origins), "NONE" (no configuration).`,
					},
					"interface": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: `Network interface to listen on. Used with listening_mode "SINGLE" or "BIND".`,
					},
					"reply_when_busy": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("ALLOW"),
						Description: `How to handle queries when the gravity database is busy. Valid values: "ALLOW", "BLOCK", "REFUSE", "DROP".`,
					},
					"block_ttl": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(2),
						Description: "TTL in seconds for blocked DNS responses.",
					},
					"blocking": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							blockingAttrTypes,
							map[string]attr.Value{
								"active": types.BoolValue(true),
								"mode":   types.StringValue("NULL"),
								"edns":   types.StringValue("TEXT"),
							},
						)),
						Attributes: map[string]schema.Attribute{
							"active": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "Whether DNS blocking is active.",
							},
							"mode": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("NULL"),
								Description: `Blocking mode: "NULL", "NXDOMAIN", "NODATA", "IP", or "IP-NODATA-AAAA".`,
							},
							"edns": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("TEXT"),
								Description: `EDNS info in blocked replies. Valid values: "NONE", "CODE", "TEXT".`,
							},
						},
					},
					"rate_limit": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							rateLimitAttrTypes,
							map[string]attr.Value{
								"count":    types.Int64Value(1000),
								"interval": types.Int64Value(60),
							},
						)),
						Attributes: map[string]schema.Attribute{
							"count": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(1000),
								Description: "Number of queries allowed per interval before rate-limiting a client.",
							},
							"interval": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(60),
								Description: "Rate-limit interval in seconds.",
							},
						},
					},
					"cache": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							cacheAttrTypes,
							map[string]attr.Value{
								"size":      types.Int64Value(10000),
								"optimizer": types.Int64Value(3600),
							},
						)),
						Attributes: map[string]schema.Attribute{
							"size": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(10000),
								Description: "DNS cache size (number of entries).",
							},
							"optimizer": schema.Int64Attribute{
								Optional:    true,
								Computed:    true,
								Default:     int64default.StaticInt64(3600),
								Description: "Cache optimizer TTL in seconds.",
							},
						},
					},
					"special_domains": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							specialDomainsAttrTypes,
							map[string]attr.Value{
								"mozilla_canary":       types.BoolValue(true),
								"icloud_private_relay": types.BoolValue(true),
								"designated_resolver":  types.BoolValue(true),
							},
						)),
						Attributes: map[string]schema.Attribute{
							"mozilla_canary": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "Block Mozilla's canary domain to disable DNS-over-HTTPS in Firefox.",
							},
							"icloud_private_relay": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "Block Apple iCloud Private Relay.",
							},
							"designated_resolver": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "Block DNS Designated Resolver records.",
							},
						},
					},
				},
			},
			"webserver": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					webserverAttrTypes,
					map[string]attr.Value{
						"domain": types.StringValue("pi.hole"),
						"interface": types.ObjectValueMust(interfaceAttrTypes, map[string]attr.Value{
							"theme": types.StringValue("default-auto"),
							"boxed": types.BoolValue(true),
						}),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"domain": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("pi.hole"),
						Description: `Hostname the web server redirects to, e.g. "pihole.example". Defaults to "pi.hole".`,
					},
					"interface": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						Default: objectdefault.StaticValue(types.ObjectValueMust(
							interfaceAttrTypes,
							map[string]attr.Value{
								"theme": types.StringValue("default-auto"),
								"boxed": types.BoolValue(true),
							},
						)),
						Attributes: map[string]schema.Attribute{
							"theme": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Default:     stringdefault.StaticString("default-auto"),
								Description: `Web interface theme. Valid values: "default-auto" (Pi-hole auto), "default-light" (Pi-hole day), "default-dark" (Pi-hole midnight), "default-darker" (Pi-hole deep-midnight), "high-contrast" (High-contrast light), "high-contrast-dark" (High-contrast dark), "lcars" (Star Trek LCARS).`,
							},
							"boxed": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "Use boxed layout for the web interface.",
							},
						},
					},
				},
			},
			"misc": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					miscAttrTypes,
					map[string]attr.Value{
						"privacy_level": types.Int64Value(0),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"privacy_level": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Privacy level for statistics. 0=full, 1=hide domains, 2=hide domains+clients, 3=anonymous.",
					},
				},
			},
			"database": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					databaseAttrTypes,
					map[string]attr.Value{
						"max_db_days": types.Int64Value(91),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"max_db_days": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(91),
						Description: "How many days to retain queries in the database. Set to 0 to disable.",
					},
				},
			},
		},
	}
}

func (r *ConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*pihole.Client)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", "expected *pihole.Client")
		return
	}
	r.client = client
}

func (r *ConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := configFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateServerConfig(cfg); err != nil {
		resp.Diagnostics.AddError("failed to apply config", err.Error())
		return
	}

	plan.ID = types.StringValue("config")
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetServerConfig()
	if err != nil {
		resp.Diagnostics.AddError("failed to read config", err.Error())
		return
	}

	upstreamVals := make([]attr.Value, len(cfg.DNS.Upstreams))
	for i, u := range cfg.DNS.Upstreams {
		upstreamVals[i] = types.StringValue(u)
	}
	upstreams, diags := types.ListValue(types.StringType, upstreamVals)
	resp.Diagnostics.Append(diags...)

	blocking, diags := types.ObjectValueFrom(ctx, blockingAttrTypes, blockingModel{
		Active: types.BoolValue(cfg.DNS.Blocking.Active),
		Mode:   types.StringValue(cfg.DNS.Blocking.Mode),
		EDNS:   types.StringValue(cfg.DNS.Blocking.EDNS),
	})
	resp.Diagnostics.Append(diags...)

	rateLimit, diags := types.ObjectValueFrom(ctx, rateLimitAttrTypes, rateLimitModel{
		Count:    types.Int64Value(int64(cfg.DNS.RateLimit.Count)),
		Interval: types.Int64Value(int64(cfg.DNS.RateLimit.Interval)),
	})
	resp.Diagnostics.Append(diags...)

	cache, diags := types.ObjectValueFrom(ctx, cacheAttrTypes, cacheModel{
		Size:      types.Int64Value(int64(cfg.DNS.Cache.Size)),
		Optimizer: types.Int64Value(int64(cfg.DNS.Cache.Optimizer)),
	})
	resp.Diagnostics.Append(diags...)

	specialDomains, diags := types.ObjectValueFrom(ctx, specialDomainsAttrTypes, specialDomainsModel{
		MozillaCanary:      types.BoolValue(cfg.DNS.SpecialDomains.MozillaCanary),
		ICloudPrivateRelay: types.BoolValue(cfg.DNS.SpecialDomains.ICloudPrivateRelay),
		DesignatedResolver: types.BoolValue(cfg.DNS.SpecialDomains.DesignatedResolver),
	})
	resp.Diagnostics.Append(diags...)

	dns, diags := types.ObjectValueFrom(ctx, dnsAttrTypes, dnsModel{
		Upstreams:        upstreams,
		DomainNeeded:     types.BoolValue(cfg.DNS.DomainNeeded),
		ExpandHosts:      types.BoolValue(cfg.DNS.ExpandHosts),
		BogusPriv:        types.BoolValue(cfg.DNS.BogusPriv),
		DNSSEC:           types.BoolValue(cfg.DNS.DNSSEC),
		QueryLogging:     types.BoolValue(cfg.DNS.QueryLogging),
		CNAMEDeepInspect: types.BoolValue(cfg.DNS.CNAMEDeepInspect),
		BlockESNI:        types.BoolValue(cfg.DNS.BlockESNI),
		HostRecord:       types.StringValue(cfg.DNS.HostRecord),
		ListeningMode:    types.StringValue(cfg.DNS.ListeningMode),
		Interface:        types.StringValue(cfg.DNS.Interface),
		ReplyWhenBusy:    types.StringValue(cfg.DNS.ReplyWhenBusy),
		BlockTTL:         types.Int64Value(int64(cfg.DNS.BlockTTL)),
		Blocking:         blocking,
		RateLimit:        rateLimit,
		Cache:            cache,
		SpecialDomains:   specialDomains,
	})
	resp.Diagnostics.Append(diags...)

	iface, diags := types.ObjectValueFrom(ctx, interfaceAttrTypes, interfaceModel{
		Theme: types.StringValue(cfg.Theme),
		Boxed: types.BoolValue(cfg.Boxed),
	})
	resp.Diagnostics.Append(diags...)

	webserver, diags := types.ObjectValueFrom(ctx, webserverAttrTypes, webserverModel{
		Domain:    types.StringValue(cfg.WebserverDomain),
		Interface: iface,
	})
	resp.Diagnostics.Append(diags...)

	misc, diags := types.ObjectValueFrom(ctx, miscAttrTypes, miscModel{
		PrivacyLevel: types.Int64Value(int64(cfg.PrivacyLevel)),
	})
	resp.Diagnostics.Append(diags...)

	database, diags := types.ObjectValueFrom(ctx, databaseAttrTypes, databaseModel{
		MaxDBDays: types.Int64Value(int64(cfg.MaxDBDays)),
	})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.DNS = dns
	state.Webserver = webserver
	state.Misc = misc
	state.Database = database
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := configFromPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateServerConfig(cfg); err != nil {
		resp.Diagnostics.AddError("failed to update config", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *ConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Config is not deleted, just removed from Terraform management.
}

func configFromPlan(ctx context.Context, plan ConfigResourceModel) (*pihole.ServerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	var dns dnsModel
	diags.Append(plan.DNS.As(ctx, &dns, basetypes.ObjectAsOptions{})...)

	var blocking blockingModel
	diags.Append(dns.Blocking.As(ctx, &blocking, basetypes.ObjectAsOptions{})...)

	var rateLimit rateLimitModel
	diags.Append(dns.RateLimit.As(ctx, &rateLimit, basetypes.ObjectAsOptions{})...)

	var cache cacheModel
	diags.Append(dns.Cache.As(ctx, &cache, basetypes.ObjectAsOptions{})...)

	var specialDomains specialDomainsModel
	diags.Append(dns.SpecialDomains.As(ctx, &specialDomains, basetypes.ObjectAsOptions{})...)

	var webserver webserverModel
	diags.Append(plan.Webserver.As(ctx, &webserver, basetypes.ObjectAsOptions{})...)

	var iface interfaceModel
	diags.Append(webserver.Interface.As(ctx, &iface, basetypes.ObjectAsOptions{})...)

	var misc miscModel
	diags.Append(plan.Misc.As(ctx, &misc, basetypes.ObjectAsOptions{})...)

	var database databaseModel
	diags.Append(plan.Database.As(ctx, &database, basetypes.ObjectAsOptions{})...)

	var upstreams []string
	diags.Append(dns.Upstreams.ElementsAs(ctx, &upstreams, false)...)

	hostRecord := dns.HostRecord.ValueString()
	piholePTR := "PI.HOLE"
	if hostRecord != "" {
		piholePTR = "NONE"
	}

	cfg := &pihole.ServerConfig{
		DNS: pihole.DNSConfig{
			Upstreams:        upstreams,
			DomainNeeded:     dns.DomainNeeded.ValueBool(),
			ExpandHosts:      dns.ExpandHosts.ValueBool(),
			BogusPriv:        dns.BogusPriv.ValueBool(),
			DNSSEC:           dns.DNSSEC.ValueBool(),
			QueryLogging:     dns.QueryLogging.ValueBool(),
			CNAMEDeepInspect: dns.CNAMEDeepInspect.ValueBool(),
			BlockESNI:        dns.BlockESNI.ValueBool(),
			HostRecord:       hostRecord,
			PiholePTR:        piholePTR,
			ListeningMode:    dns.ListeningMode.ValueString(),
			Interface:        dns.Interface.ValueString(),
			ReplyWhenBusy:    dns.ReplyWhenBusy.ValueString(),
			BlockTTL:         int(dns.BlockTTL.ValueInt64()),
			Blocking: pihole.DNSBlocking{
				Active: blocking.Active.ValueBool(),
				Mode:   blocking.Mode.ValueString(),
				EDNS:   blocking.EDNS.ValueString(),
			},
			RateLimit: pihole.DNSRateLimit{
				Count:    int(rateLimit.Count.ValueInt64()),
				Interval: int(rateLimit.Interval.ValueInt64()),
			},
			Cache: pihole.DNSCache{
				Size:      int(cache.Size.ValueInt64()),
				Optimizer: int(cache.Optimizer.ValueInt64()),
			},
			SpecialDomains: pihole.DNSSpecialDomains{
				MozillaCanary:      specialDomains.MozillaCanary.ValueBool(),
				ICloudPrivateRelay: specialDomains.ICloudPrivateRelay.ValueBool(),
				DesignatedResolver: specialDomains.DesignatedResolver.ValueBool(),
			},
		},
		WebserverDomain: webserver.Domain.ValueString(),
		Theme:           iface.Theme.ValueString(),
		Boxed:           iface.Boxed.ValueBool(),
		PrivacyLevel:    int(misc.PrivacyLevel.ValueInt64()),
		MaxDBDays:       int(database.MaxDBDays.ValueInt64()),
	}

	return cfg, diags
}
