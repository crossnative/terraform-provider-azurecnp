package provider

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &subscriptionPoolResource{}
	_ resource.ResourceWithConfigure = &subscriptionPoolResource{}
)

// NewSubscriptionPoolResource is a helper function to simplify the provider implementation.
func NewSubscriptionPoolResource() resource.Resource {
	return &subscriptionPoolResource{}
}

// subscriptionPoolResource is the resource implementation.
type subscriptionPoolResource struct {
	clientFactory *armmanagementgroups.ClientFactory
}

type subscriptionPoolResourceModel struct {
	PoolManagementGroupName     types.String `tfsdk:"pool_management_group_name"`
	TargetManagementGroupName   types.String `tfsdk:"target_management_group_name"`
	SubscriptionId              types.String `tfsdk:"subscription_id"`
	NewSubscriptionName         types.String `tfsdk:"new_subscription_name"`
	ActualSubscriptionName      types.String `tfsdk:"actual_subscription_name"`
	ActualParentManagementGroup types.String `tfsdk:"actual_parant_management_group"`
}

// Metadata returns the resource type name.
func (r *subscriptionPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_pool"
}

// Schema defines the schema for the resource.
func (r *subscriptionPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pool_management_group_name": schema.StringAttribute{
				Required: true,
			},
			"target_management_group_name": schema.StringAttribute{
				Required: true,
			},
			"subscription_id": schema.StringAttribute{
				Required: true,
			},
			"new_subscription_name": schema.StringAttribute{
				Required: true,
			},
			"actual_subscription_name": schema.StringAttribute{
				Computed: true,
			},
			"actual_parant_management_group": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *subscriptionPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	clientFactory, ok := req.ProviderData.(*armmanagementgroups.ClientFactory)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.clientFactory = clientFactory
}

// Create a new resource.
func (r *subscriptionPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan subscriptionPoolResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO check if subscription is in pool

	// Associate Subscription
	response, err := r.clientFactory.NewManagementGroupSubscriptionsClient().Create(context.Background(), plan.TargetManagementGroupName.ValueString(), plan.SubscriptionId.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error moving subscription", err.Error(),
		)
		return
	}
	plan.ActualSubscriptionName = types.StringValue(*response.Properties.DisplayName)
	plan.ActualParentManagementGroup = plan.TargetManagementGroupName

	//TODO rename subscription

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subscriptionPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state subscriptionPoolResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	var isInPool bool = false
	response, err := r.clientFactory.NewManagementGroupSubscriptionsClient().GetSubscription(context.Background(), state.PoolManagementGroupName.ValueString(), state.SubscriptionId.ValueString(), nil)
	if err == nil {
		//TODO actually check if its a 404 to be sure
		isInPool = true
		state.ActualSubscriptionName = types.StringValue(*response.Properties.DisplayName)
	}

	var isInTarget bool = false
	if !isInPool {
		response, err := r.clientFactory.NewManagementGroupSubscriptionsClient().GetSubscription(context.Background(), state.TargetManagementGroupName.ValueString(), state.SubscriptionId.ValueString(), nil)
		if err == nil {
			isInTarget = true
			state.ActualSubscriptionName = types.StringValue(*response.Properties.DisplayName)
		}
	}

	if (!isInPool) && (!isInTarget) {
		resp.Diagnostics.AddError(
			"Subscription neither in pool nor in target",
			"Broken State. Subscription should be either in pool or in target.",
		)
		return
	}

	if isInPool {
		state.ActualParentManagementGroup = state.PoolManagementGroupName
	}
	if isInTarget {
		state.ActualParentManagementGroup = state.TargetManagementGroupName
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *subscriptionPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subscriptionPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
