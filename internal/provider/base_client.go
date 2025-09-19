package provider

import (
	"context"

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

func (b BaseClient) RenameSubscription(subscriptionId string, name string) (armsubscription.ClientRenameResponse, error) {
	return b.subscriptionClientFactory.NewClient().Rename(context.Background(), subscriptionId, armsubscription.Name{SubscriptionName: &name}, nil)
}

func (b BaseClient) MoveSubscription(subscriptionId string, managementGroupId string) (armmanagementgroups.ManagementGroupSubscriptionsClientCreateResponse, error) {
	return b.managementGroupClientFactory.NewManagementGroupSubscriptionsClient().Create(context.Background(), managementGroupId, subscriptionId, nil)
}

func (b BaseClient) ReadSubscriptionState(subscriptionId string) (*armmanagementgroups.EntityInfo, error) {
	pager := b.managementGroupClientFactory.NewEntitiesClient().NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}
		for _, entityInfo := range page.Value {
			if *entityInfo.Type == "/subscriptions" && *entityInfo.Name == subscriptionId {
				return entityInfo, nil
			}
		}
	}
	return nil, NewNoSubscriptionsFoundError(subscriptionId)
}
