package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

type ListResource struct {
	client *pihole.Client
}

type ListResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
	Type    types.String `tfsdk:"type"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Comment types.String `tfsdk:"comment"`
	Groups  types.Set    `tfsdk:"groups"`
}

func NewListResource() resource.Resource { return &ListResource{} }

func (r *ListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (r *ListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole block or allow list.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "URL of the list.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: `List type: "block" or "allow".`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the list is enabled.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional comment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"groups": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default: setdefault.StaticValue(
					types.SetValueMust(types.StringType, []attr.Value{types.StringValue("Default")}),
				),
				Description: "Group names to assign this list to.",
			},
		},
	}
}

func (r *ListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	list, err := r.client.CreateList(pihole.List{
		Address: plan.Address.ValueString(),
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Comment: plan.Comment.ValueString(),
		Groups:  groupIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create list", err.Error())
		return
	}

	plan.ID = plan.Address
	plan.Comment = types.StringValue(list.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.client.GetListByAddress(state.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read list", err.Error())
		return
	}
	if list == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	groupNames, err := r.client.ResolveGroupIDs(list.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve group IDs", err.Error())
		return
	}

	groupVals := make([]attr.Value, len(groupNames))
	for i, n := range groupNames {
		groupVals[i] = types.StringValue(n)
	}
	groups, diags := types.SetValue(types.StringType, groupVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Type = types.StringValue(list.Type)
	state.Enabled = types.BoolValue(list.Enabled)
	state.Comment = types.StringValue(list.Comment)
	state.Groups = groups
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.client.GetListByAddress(plan.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to look up list for update", err.Error())
		return
	}
	if list == nil {
		resp.Diagnostics.AddError("list not found", plan.Address.ValueString())
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	err = r.client.UpdateList(list.ID, pihole.List{
		Address: plan.Address.ValueString(),
		Type:    plan.Type.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Comment: plan.Comment.ValueString(),
		Groups:  groupIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update list", err.Error())
		return
	}

	updated, err := r.client.GetListByAddress(plan.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read list after update", err.Error())
		return
	}
	if updated != nil {
		plan.Comment = types.StringValue(updated.Comment)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.client.GetListByAddress(state.Address.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to look up list for delete", err.Error())
		return
	}
	if list == nil {
		return
	}

	if err := r.client.DeleteList(list.Address, list.Type); err != nil {
		resp.Diagnostics.AddError("failed to delete list", err.Error())
	}
}

func (r *ListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("address"), req.ID)...)
}

func (r *ListResource) resolveGroups(ctx context.Context, groups types.Set) ([]int, error) {
	var names []string
	diags := groups.ElementsAs(ctx, &names, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract group names")
	}
	return r.client.ResolveGroupNames(names)
}
