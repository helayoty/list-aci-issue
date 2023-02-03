package client

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	azaciv2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	location = "eastus2"
)

type Config struct {
	ClientID             string
	UserIdentityClientId string
	TenantID             string
	SubscriptionID       string
	CGClient             *azaciv2.ContainerGroupsClient
	RGClient             *armresources.ResourceGroupsClient
}

func (config *Config) InitClients() {
	if config.SubscriptionID == "" {
		log.Println("subscriptionID cannot be empty")
		os.Exit(1)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	client, err := azaciv2.NewContainerGroupsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create CG client: %v", err)
	}
	config.CGClient = client

	resourceGroupClient, err := armresources.NewResourceGroupsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create RG client: %v", err)
	}
	config.RGClient = resourceGroupClient

	log.Println("clients are ready")
}

func (config *Config) GetContainerGroupList(ctx context.Context, rgName string) {

	pager := config.CGClient.NewListByResourceGroupPager(rgName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, cg := range nextResult.Value {
			ipaddress := "nil"
			if cg.Properties.IPAddress != nil {
				ipaddress = *cg.Properties.IPAddress.IP
			}
			cgInstanceView := azaciv2.ContainerGroupPropertiesInstanceView{}
			if cg.Properties.InstanceView != nil {
				cgInstanceView = *cg.Properties.InstanceView
			}

			containerInstanceView := azaciv2.ContainerPropertiesInstanceView{}
			if cg.Properties.Containers[0].Properties.InstanceView != nil {
				containerInstanceView = *cg.Properties.Containers[0].Properties.InstanceView
			}
			log.Printf("%s info: \ncommands: %v\niamge: %v\nCPU resource: %v\nprovisingState: %v\nOSType:%v\nSKU: %v\nIPAddress:  %v\nInstanceView: %v\nContainer InstanceView: %v\n\n", *cg.Name, *cg.Properties.Containers[0].Properties.Command[0],
				*cg.Properties.Containers[0].Properties.Image,
				*cg.Properties.Containers[0].Properties.Resources.Requests.CPU,
				*cg.Properties.ProvisioningState,
				*cg.Properties.OSType,
				*cg.Properties.SKU,
				ipaddress,
				cgInstanceView,
				containerInstanceView)
		}
	}
}

func (config *Config) CreateContainerGroup(ctx context.Context, cgName, rgName string) {
	poller, err := config.CGClient.BeginCreateOrUpdate(ctx, rgName, cgName, azaciv2.ContainerGroup{
		Properties: &azaciv2.ContainerGroupPropertiesProperties{
			Containers: []*azaciv2.Container{
				{
					Name: to.Ptr(cgName),
					Properties: &azaciv2.ContainerProperties{
						Command: []*string{
							to.Ptr("/bin/sh"),
							to.Ptr("-c"),
							to.Ptr("sleep 10")},
						Image: to.Ptr("alpine:latest"),
						Resources: &azaciv2.ResourceRequirements{
							Requests: &azaciv2.ResourceRequests{
								CPU:        to.Ptr[float64](1),
								MemoryInGB: to.Ptr[float64](1),
							},
						},
					},
				}},
			OSType:        to.Ptr(azaciv2.OperatingSystemTypesLinux),
			RestartPolicy: to.Ptr(azaciv2.ContainerGroupRestartPolicyNever),
			SKU:           to.Ptr(azaciv2.ContainerGroupSKUStandard),
		},
		Location: to.Ptr(location),
	}, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}
	// TODO: use response item
	_ = res
}

func (config *Config) DeleteContainerGroup(ctx context.Context, rgName, cgName string) {

	poller, err := config.CGClient.BeginDelete(ctx, rgName, cgName, nil)
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to pull the result: %v", err)
	}
	// TODO: use response item
	_ = res
}

func (config *Config) CheckExistenceResourceGroup(ctx context.Context, rgName string) bool {
	boolResp, err := config.RGClient.CheckExistence(ctx, rgName, nil)
	if err != nil {
		log.Fatalf("failed to check rg: %s", rgName)
	}
	return boolResp.Success
}

func (config *Config) CreateResourceGroup(ctx context.Context, rgName string) *armresources.ResourceGroup {
	resourceGroupResp, err := config.RGClient.CreateOrUpdate(
		ctx,
		rgName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		log.Fatalf("failed to create rg: %s", rgName)
	}
	return &resourceGroupResp.ResourceGroup
}

func (config *Config) DeleteResourceGroup(ctx context.Context, rgName string) {
	pollerResp, err := config.RGClient.BeginDelete(ctx, rgName, nil)
	if err != nil {
		log.Fatalf("failed to deleting rg: %s", rgName)
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		log.Fatalf("failed to deleting rg: %s", rgName)
	}
}
