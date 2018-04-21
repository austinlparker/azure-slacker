package main

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"os"
)

//GetAzureSubscription returns the subscription id from env vars
func GetAzureSubscription() string {
	return os.Getenv("AZURE_SUBSCRIPTION_ID")
}

//GetSlackWebhook returns the Slack webhook url from env vars
func GetSlackWebhook() string {
	return os.Getenv("SLACK_WEBHOOK_URL")
}

//GetGroupByPrefix returns the string prefix to group virtual machines from env vars
func GetGroupByPrefix() string {
	return os.Getenv("AZURE_GROUP_BY_PREFIX")
}

//GetResourceGroupName returns the resource group from env vars
func GetResourceGroupName() string {
	return os.Getenv("AZURE_RESOURCE_GROUP_NAME")
}

//GetVMClient returns an Azure VirtualMachinesClient using env vars for authorization
func GetVMClient() compute.VirtualMachinesClient {
	vmClient := compute.NewVirtualMachinesClient(GetAzureSubscription())
	auth, _ := auth.NewAuthorizerFromEnvironment()
	vmClient.Authorizer = auth
	return vmClient
}
