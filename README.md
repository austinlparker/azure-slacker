# Azure Environment Status To Slack

This software is intended to allow for grouping and posting Azure 'Environments' to a Slack channel.

An 'environment' is a collection of Azure VM's contained in the same ARM (Azure Resource Manager) group that share a specified prefix.
Essentially, consider if your resource group contains multiple virtual machines, with several of them related through a tag like 'Type'. Perhaps there's several different staging or development environments running, each with several machines. This program will help in displaying the status of these machines and posting it to a Slack Webhook.

The same principal can be applied to posting this data to other webhooks, or out to a console, with some minimal modifications.

## Environment Variables

There are a variety of environment variables used for configuration, they are as follows:

* AZURE_TENANT_ID
* AZURE_CLIENT_ID
* AZURE_CLIENT_SECRET
* AZURE_SUBSCRIPTION_ID
* AZURE_GROUP_BY_PREFIX
* AZURE_RESOURCE_GROUP_NAME
* SLACK_WEBHOOK

Creating these values is slightly outside the scope of this README, although the following sources are helpful for configuring an Azure Active Directory application with RBAC.

[Creating Azure AD Application](https://docs.microsoft.com/en-us/azure/azure-resource-manager/resource-group-create-service-principal-portal) - This will give you the tenant and client identifiers, as well as the client secret (this is referred to as a key in the Azure documentation).

Slack webhooks can be created from your Slack team's customization settings.

Take special note of the `AZURE_GROUP_BY_PREFIX` value. This is the tag key that contains the value you wish to group results by. So, for example, if I wanted to group all virtual machines in my resource group by the key 'Type' (where values might be something like 'Staging1', 'Staging2', etc.) then I would set this environment variable to 'Type'.

## Developing and Deploying

Tested with Go 1.10.

Install dependencies using `dep ensure`.
Build with `go build main.go helpers.go`

You can also utilize the included `Dockerfile` to build and run the software; `docker build -t azwatch .` and `docker run -it -rm --name azwatcher azwatch`.
