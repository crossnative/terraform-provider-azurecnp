package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrderResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "azurecnp_subscription_pool_lease" "test" {
  pool_management_group_name = "Crossnative"
  pool_subscription_prefix = "Azure_Subscription_Crossnative_Pool_"
  target_management_group_name = "cn-sandbox"
  target_subscription_name = "automatic-test-create"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.#", "1"),
					// Verify first order item
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.pool_management_group_name", "Crossnative"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.pool_subscription_prefix", "Azure_Subscription_Crossnative_Pool_"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.target_management_group_name", "cn-sandbox"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.actual_parant_management_group", "cn-sandbox"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.target_subscription_name", "automatic-test-create"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("azurecnp_subscription_pool_lease.test", "subscription_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "azurecnp_subscription_pool_lease.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "azurecnp_subscription_pool_lease" "test" {
  pool_management_group_name = "Crossnative"
  pool_subscription_prefix = "Azure_Subscription_Crossnative_Pool_"
  target_management_group_name = "cn-hosting"
  target_subscription_name = "automatic-test-update"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.pool_management_group_name", "Crossnative"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.pool_subscription_prefix", "Azure_Subscription_Crossnative_Pool_"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.target_management_group_name", "cn-hosting"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.actual_parant_management_group", "cn-hosting"),
					resource.TestCheckResourceAttr("azurecnp_subscription_pool_lease.test", "items.0.target_subscription_name", "automatic-test-update"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("azurecnp_subscription_pool_lease.test", "subscription_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
