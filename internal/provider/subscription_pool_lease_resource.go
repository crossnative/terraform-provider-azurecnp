package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &subscriptionPoolLeaseResource{}
	_ resource.ResourceWithConfigure   = &subscriptionPoolLeaseResource{}
	_ resource.ResourceWithImportState = &subscriptionPoolLeaseResource{}
)

// NewSubscriptionPoolResource is a helper function to simplify the provider implementation.
func NewSubscriptionPoolLeaseResource() resource.Resource {
	return &subscriptionPoolLeaseResource{}
}

// subscriptionPoolLeaseResource is the resource implementation.
type subscriptionPoolLeaseResource struct {
	clientFactoryHolder *ClientFactoryHolder
}

type subscriptionPoolLeaseResourceModel struct {
	PoolManagementGroupName      types.String `tfsdk:"pool_management_group_name"`
	PoolSubscriptionPrefix       types.String `tfsdk:"pool_subscription_prefix"`
	TargetManagementGroupName    types.String `tfsdk:"target_management_group_name"`
	TargetSubscriptionName       types.String `tfsdk:"target_subscription_name"`
	SubscriptionId               types.String `tfsdk:"subscription_id"`
	FullyQualifiedSubscriptionId types.String `tfsdk:"fully_qualified_subscription_id"`
	ActualParentManagementGroup  types.String `tfsdk:"actual_parant_management_group"`
}

// Metadata returns the resource type name.
func (r *subscriptionPoolLeaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_pool_lease"
}

// Schema defines the schema for the resource.
func (r *subscriptionPoolLeaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pool_management_group_name": schema.StringAttribute{
				Required: true,
			},
			"pool_subscription_prefix": schema.StringAttribute{
				Required: true,
			},
			"target_management_group_name": schema.StringAttribute{
				Required: true,
			},
			"target_subscription_name": schema.StringAttribute{
				Required: true,
			},
			"subscription_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fully_qualified_subscription_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"actual_parant_management_group": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *subscriptionPoolLeaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	clientFactory, ok := req.ProviderData.(*ClientFactoryHolder)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *provider.ClientFactoryHolder, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.clientFactoryHolder = clientFactory
}

// Create a new resource.
func (r *subscriptionPoolLeaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan subscriptionPoolLeaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := plan.PoolSubscriptionPrefix.ValueString()

	subscriptionPager := r.clientFactoryHolder.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().NewGetSubscriptionsUnderManagementGroupPager(plan.PoolManagementGroupName.ValueString(), nil)
	var matchingSubscription *armmanagementgroups.SubscriptionUnderManagementGroup
root:
	for subscriptionPager.More() {
		page, err := subscriptionPager.NextPage(context.Background())
		if err != nil {
			resp.Diagnostics.AddError("Failed fetching Subscriptions from Pool", err.Error())
			return
		}
		for _, sub := range page.Value {
			if strings.HasPrefix(*sub.Properties.DisplayName, prefix) {
				matchingSubscription = sub
				break root
			}
		}
	}
	if matchingSubscription == nil {
		resp.Diagnostics.AddError(
			"Didn't find any available Subscription",
			fmt.Sprintf("Searched for subscriptions with prefix '%s' in ManagementGroup '%s'", prefix, plan.PoolManagementGroupName.ValueString()),
		)
		return
	}

	// Associate Subscription
	_, err := r.clientFactoryHolder.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().Create(context.Background(), plan.TargetManagementGroupName.ValueString(), *matchingSubscription.Name, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error moving subscription", err.Error(),
		)
		return
	}
	plan.ActualParentManagementGroup = types.StringValue(plan.TargetManagementGroupName.ValueString())

	_, err = r.clientFactoryHolder.subscriptionClientFactory.NewClient().Rename(context.Background(), *matchingSubscription.Name, armsubscription.Name{SubscriptionName: plan.TargetSubscriptionName.ValueStringPointer()}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error renaming subscription", err.Error(),
		)
		return
	}
	plan.SubscriptionId = types.StringValue(*matchingSubscription.Name)
	plan.FullyQualifiedSubscriptionId = types.StringValue(*matchingSubscription.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subscriptionPoolLeaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state subscriptionPoolLeaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pager := r.clientFactoryHolder.managementGroupClientFactory.NewEntitiesClient().NewListPager(nil)
	var matchingEntity *armmanagementgroups.EntityInfo
root:
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			resp.Diagnostics.AddError("Failed fetching entities", err.Error())
			return
		}
		for _, entityInfo := range page.Value {
			if *entityInfo.Type == "/subscriptions" && *entityInfo.Name == state.SubscriptionId.ValueString() {
				matchingEntity = entityInfo
				break root
			}
		}
	}
	if matchingEntity == nil {
		resp.Diagnostics.AddError(
			"Couldn't find managed subscription",
			fmt.Sprintf("Couldn't find managed subscription: %s", state.SubscriptionId.ValueString()),
		)
		return
	}
	state.ActualParentManagementGroup = types.StringValue(strings.TrimPrefix(*matchingEntity.Properties.Parent.ID, "/providers/Microsoft.Management/managementGroups/"))
	state.TargetSubscriptionName = types.StringValue(*matchingEntity.Properties.DisplayName)
	state.SubscriptionId = types.StringValue(*matchingEntity.Name)
	state.FullyQualifiedSubscriptionId = types.StringValue(*matchingEntity.ID)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *subscriptionPoolLeaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan subscriptionPoolLeaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state subscriptionPoolLeaseResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.clientFactoryHolder.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().GetSubscription(context.Background(), state.ActualParentManagementGroup.ValueString(), state.SubscriptionId.ValueString(), nil)
	if err != nil {
		//TODO check for 404 or differente error
		resp.Diagnostics.AddError(
			"Broken State",
			fmt.Sprintf("Could not find Subscription '%s' under ManagementGroup '%s'\nAzure API Error: %s", state.SubscriptionId.ValueString(), state.ActualParentManagementGroup.ValueString(), err.Error()),
		)
		return
	}

	if state.ActualParentManagementGroup.ValueString() != plan.TargetManagementGroupName.ValueString() {
		_, err := r.clientFactoryHolder.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().Create(ctx, plan.TargetManagementGroupName.ValueString(), *sub.Name, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error during Subscription Move",
				err.Error(),
			)
			return
		}
	}
	plan.ActualParentManagementGroup = types.StringValue(plan.TargetManagementGroupName.ValueString())

	if plan.TargetSubscriptionName.ValueString() != *sub.Properties.DisplayName {
		_, err := r.clientFactoryHolder.subscriptionClientFactory.NewClient().Rename(context.Background(), state.SubscriptionId.ValueString(), armsubscription.Name{SubscriptionName: plan.TargetSubscriptionName.ValueStringPointer()}, nil)
		if err != nil {
			//TODO handle retry on 429
			resp.Diagnostics.AddError(
				"Error during Subscription Rename",
				err.Error(),
			)
			return
		}
		plan.TargetSubscriptionName = types.StringValue(plan.TargetSubscriptionName.ValueString())
	}
	plan.SubscriptionId = types.StringValue(*sub.Name)
	plan.FullyQualifiedSubscriptionId = types.StringValue(*sub.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subscriptionPoolLeaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subscriptionPoolLeaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.clientFactoryHolder.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().Create(context.Background(), state.PoolManagementGroupName.ValueString(), state.SubscriptionId.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error during Subscription Move",
			err.Error(),
		)
		return
	}

	newSubscriptionName := truncateString(state.PoolSubscriptionPrefix.ValueString()+state.SubscriptionId.ValueString(), 64)
	_, err = r.clientFactoryHolder.subscriptionClientFactory.NewClient().Rename(context.Background(), state.SubscriptionId.ValueString(), armsubscription.Name{SubscriptionName: &newSubscriptionName}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error during Subscription Rename",
			err.Error(),
		)
		return
	}
}

func (r *subscriptionPoolLeaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("subscription_id"), req, resp)
}

func truncateString(s string, max int) string {
	return s[:max]
}
