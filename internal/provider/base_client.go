package provider

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

type BaseClient struct {
	managementGroupClientFactory *armmanagementgroups.ClientFactory
	subscriptionClientFactory    *armsubscription.ClientFactory
	availableSubscriptions       chan string
	poolManagementGroupId        string
	poolSubscriptionPrefix       string
}
