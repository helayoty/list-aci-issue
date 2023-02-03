package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"list-aci-issue/pkg/client"
)

var (
	rgName = "test-list-aci"
)

func enterInput(param string) string {
	log.Printf("Please enter the %s: ", param)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Println("An error occurred while reading ", param, ". Please try again", err)
		return ""
	}
	result := strings.TrimSuffix(input, "\n")
	log.Println(result)
	return result
}

func main() {
	ctx := context.Background()

	config := client.Config{}
	config.SubscriptionID = enterInput("SubscriptionID")
	if config.SubscriptionID == "" {
		log.Println("SubscriptionID input is mandatory")
		os.Exit(1)
	}

	config.InitClients()

	var resourceGroup *armresources.ResourceGroup
	exits := config.CheckExistenceResourceGroup(ctx, rgName)

	if exits {
		log.Println("resources group already exist. Skipping creating new one")
	} else {
		resourceGroup = config.CreateResourceGroup(ctx, rgName)
		log.Println("resources group created:", *resourceGroup.ID)
	}

	log.Println("Create brand new container group")
	config.CreateContainerGroup(ctx, fmt.Sprintf("cg-%d", rand.Int()), rgName)

	log.Println("Recreate container groups")
	for i := 0; i < 1; i++ {
		cgName := fmt.Sprintf("cg-%d", i)
		config.CreateContainerGroup(ctx, cgName, rgName)
		log.Printf("container group %s has been created\n", cgName)
	}

	log.Println("Calling GetContainerGroupList Once")
	config.GetContainerGroupList(ctx, rgName)

	log.Println("Calling GetContainerGroupList Twice")
	config.GetContainerGroupList(ctx, rgName)

	log.Println("Cleaning up. Deleting resource Group")
	config.DeleteResourceGroup(ctx, rgName)

}
