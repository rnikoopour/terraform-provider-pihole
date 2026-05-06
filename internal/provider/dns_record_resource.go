package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

type DNSRecordResource struct {
	client *pihole.Client
}

type DNSRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	IP       types.String `tfsdk:"ip"`
	Hostname types.String `tfsdk:"hostname"`
}

func NewDNSRecordResource() resource.Resource { return &DNSRecordResource{} }

func (r *DNSRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole custom local DNS A record.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Hostname to resolve to the IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DNSRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec := pihole.DNSRecord{IP: plan.IP.ValueString(), Hostname: plan.Hostname.ValueString()}
	if err := r.client.CreateDNSRecord(rec); err != nil {
		resp.Diagnostics.AddError("failed to create DNS record", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s %s", rec.IP, rec.Hostname))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec, err := r.client.GetDNSRecord(state.IP.ValueString(), state.Hostname.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read DNS record", err.Error())
		return
	}
	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: "ip hostname"
	parts := strings.SplitN(req.ID, " ", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", `expected format "ip hostname"`)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostname"), parts[1])...)
}

func (r *DNSRecordResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// ip and hostname both require replace, so Update is never called.
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec := pihole.DNSRecord{IP: state.IP.ValueString(), Hostname: state.Hostname.ValueString()}
	if err := r.client.DeleteDNSRecord(rec); err != nil {
		resp.Diagnostics.AddError("failed to delete DNS record", err.Error())
	}
}
