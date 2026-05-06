package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

type GroupResource struct {
	client *pihole.Client
}

type GroupResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func NewGroupResource() resource.Resource { return &GroupResource{} }

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Group name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional comment.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the group is enabled.",
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.CreateGroup(pihole.Group{
		Name:    plan.Name.ValueString(),
		Comment: plan.Comment.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create group", err.Error())
		return
	}

	plan.ID = types.StringValue(group.Name)
	plan.Comment = types.StringValue(group.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read group", err.Error())
		return
	}
	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Comment = types.StringValue(group.Comment)
	state.Enabled = types.BoolValue(group.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateGroup(plan.Name.ValueString(), pihole.Group{
		Comment: plan.Comment.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update group", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGroup(state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete group", err.Error())
	}
}
