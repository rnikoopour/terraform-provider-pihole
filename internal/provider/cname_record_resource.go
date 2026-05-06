package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

type CNAMERecordResource struct {
	client *pihole.Client
}

type CNAMERecordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	Target types.String `tfsdk:"target"`
}

func NewCNAMERecordResource() resource.Resource { return &CNAMERecordResource{} }

func (r *CNAMERecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_record"
}

func (r *CNAMERecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole local CNAME record.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The CNAME domain (alias).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required:    true,
				Description: "The target domain this CNAME points to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CNAMERecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CNAMERecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CNAMERecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec := pihole.CNAMERecord{Domain: plan.Domain.ValueString(), Target: plan.Target.ValueString()}
	if err := r.client.CreateCNAMERecord(rec); err != nil {
		resp.Diagnostics.AddError("failed to create CNAME record", err.Error())
		return
	}

	plan.ID = plan.Domain
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CNAMERecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CNAMERecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec, err := r.client.GetCNAMERecord(state.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read CNAME record", err.Error())
		return
	}
	if rec == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Target = types.StringValue(rec.Target)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CNAMERecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), req.ID)...)
}

func (r *CNAMERecordResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// domain and target both require replace, so Update is never called.
}

func (r *CNAMERecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CNAMERecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec := pihole.CNAMERecord{Domain: state.Domain.ValueString(), Target: state.Target.ValueString()}
	if err := r.client.DeleteCNAMERecord(rec); err != nil {
		resp.Diagnostics.AddError("failed to delete CNAME record", err.Error())
	}
}
