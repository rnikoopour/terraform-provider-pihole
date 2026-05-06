package provider

import (
	"context"
	"fmt"
	"strings"

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

type DomainResource struct {
	client *pihole.Client
}

type DomainResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Type    types.String `tfsdk:"type"`
	Kind    types.String `tfsdk:"kind"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Comment types.String `tfsdk:"comment"`
	Groups  types.Set    `tfsdk:"groups"`
}

func NewDomainResource() resource.Resource { return &DomainResource{} }

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole allow or deny domain entry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain or regex pattern.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: `"allow" or "deny".`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Required:    true,
				Description: `"exact" or "regex".`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the entry is enabled.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional comment.",
			},
			"groups": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default: setdefault.StaticValue(
					types.SetValueMust(types.StringType, []attr.Value{types.StringValue("Default")}),
				),
				Description: "Group names to assign this domain entry to.",
			},
		},
	}
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	d, err := r.client.CreateDomain(pihole.Domain{
		Domain:  plan.Domain.ValueString(),
		Type:    plan.Type.ValueString(),
		Kind:    plan.Kind.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Comment: plan.Comment.ValueString(),
		Groups:  groupIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create domain", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", d.Type, d.Kind, d.Domain))
	plan.Comment = types.StringValue(d.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.GetDomain(state.Domain.ValueString(), state.Type.ValueString(), state.Kind.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to read domain", err.Error())
		return
	}
	if d == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	groupNames, err := r.client.ResolveGroupIDs(d.Groups)
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

	state.Type = types.StringValue(d.Type)
	state.Kind = types.StringValue(d.Kind)
	state.Enabled = types.BoolValue(d.Enabled)
	state.Comment = types.StringValue(d.Comment)
	state.Groups = groups
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.GetDomain(plan.Domain.ValueString(), plan.Type.ValueString(), plan.Kind.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to look up domain for update", err.Error())
		return
	}
	if d == nil {
		resp.Diagnostics.AddError("domain not found", plan.Domain.ValueString())
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	err = r.client.UpdateDomain(d.ID, pihole.Domain{
		Domain:  plan.Domain.ValueString(),
		Type:    plan.Type.ValueString(),
		Kind:    plan.Kind.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
		Comment: plan.Comment.ValueString(),
		Groups:  groupIDs,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update domain", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDomain(state.Type.ValueString(), state.Kind.ValueString(), state.Domain.ValueString()); err != nil {
		resp.Diagnostics.AddError("failed to delete domain", err.Error())
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: "type/kind/domain"
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("invalid import ID", `expected format "type/kind/domain"`)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kind"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[2])...)
}

func (r *DomainResource) resolveGroups(ctx context.Context, groups types.Set) ([]int, error) {
	var names []string
	if diags := groups.ElementsAs(ctx, &names, false); diags.HasError() {
		return nil, fmt.Errorf("failed to extract group names")
	}
	return r.client.ResolveGroupNames(names)
}
