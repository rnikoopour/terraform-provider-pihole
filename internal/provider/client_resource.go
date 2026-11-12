package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rnikoopour/terraform-provider-pihole/internal/pihole"
)

// ClientResource manages a Pi-hole client record. Assigning a client to a
// non-default group is what makes group-scoped allow/deny domains apply to
// just that client instead of the whole instance.
type ClientResource struct {
	client *pihole.Client
}

type ClientResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Client  types.String `tfsdk:"client"`
	Comment types.String `tfsdk:"comment"`
	Groups  types.Set    `tfsdk:"groups"`
}

func NewClientResource() resource.Resource { return &ClientResource{} }

func (r *ClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (r *ClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pi-hole client record for group-scoped filtering. Clients not managed here belong to the Default group only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client": schema.StringAttribute{
				Required:    true,
				Description: "Client identifier: IP address, CIDR, MAC address, hostname, or interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				Description: "Group names this client belongs to. Include \"Default\" to keep instance-wide ad blocking in effect.",
			},
		},
	}
}

func (r *ClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	entry, err := r.client.CreateClient(&pihole.CreateClientInput{
		Ctx: ctx,
		Entry: pihole.ClientEntry{
			Client:  plan.Client.ValueString(),
			Comment: plan.Comment.ValueString(),
			Groups:  groupIDs,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	plan.ID = types.StringValue(entry.Entry.Client)
	plan.Comment = types.StringValue(entry.Entry.Comment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.GetClient(&pihole.GetClientInput{Ctx: ctx, Addr: state.Client.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("failed to read client", err.Error())
		return
	}
	if out.Entry == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	groupNames, err := r.client.ResolveGroupIDs(&pihole.ResolveGroupIDsInput{Ctx: ctx, IDs: out.Entry.Groups})
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve group IDs", err.Error())
		return
	}

	groupVals := make([]attr.Value, len(groupNames.Names))
	for i, n := range groupNames.Names {
		groupVals[i] = types.StringValue(n)
	}
	groups, diags := types.SetValue(types.StringType, groupVals)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Comment = types.StringValue(out.Entry.Comment)
	state.Groups = groups
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClientResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupIDs, err := r.resolveGroups(ctx, plan.Groups)
	if err != nil {
		resp.Diagnostics.AddError("failed to resolve groups", err.Error())
		return
	}

	err = r.client.UpdateClient(&pihole.UpdateClientInput{
		Ctx:  ctx,
		Addr: plan.Client.ValueString(),
		Entry: pihole.ClientEntry{
			Client:  plan.Client.ValueString(),
			Comment: plan.Comment.ValueString(),
			Groups:  groupIDs,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update client", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client"), req.ID)...)
}

func (r *ClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClientResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteClient(&pihole.DeleteClientInput{Ctx: ctx, Addr: state.Client.ValueString()}); err != nil {
		resp.Diagnostics.AddError("failed to delete client", err.Error())
	}
}

func (r *ClientResource) resolveGroups(ctx context.Context, groups types.Set) ([]int, error) {
	var names []string
	if diags := groups.ElementsAs(ctx, &names, false); diags.HasError() {
		return nil, fmt.Errorf("failed to extract group names")
	}
	out, err := r.client.ResolveGroupNames(&pihole.ResolveGroupNamesInput{Ctx: ctx, Names: names})
	if err != nil {
		return nil, err
	}
	return out.IDs, nil
}
