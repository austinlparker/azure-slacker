package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2017-03-30/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/olekukonko/tablewriter"
)

func main() {
	ctx := context.Background()
	vmClient := GetVMClient()

	// get all virtual machines and group them by a tag
	virtualMachines := getAllVirtualMachines(ctx, vmClient)
	groupedVirtualMachines := groupVirtualMachinesByExtendedInfo(ctx, vmClient, virtualMachines)

	// create a 2d array to save our results to
	resultData := make([][]string, len(groupedVirtualMachines))
	for key, value := range groupedVirtualMachines {
		var createdAt *date.Time
		var stateChangedAt *date.Time
		var allocationCount int
		var trimmedPowerState string

		for _, item := range value {
			// gets the first attached disk
			disk := (*item.InstanceView.Disks)[0]
			// zeroeth item in this field is the provisioning status, generally?
			instanceProvisioningStatus := (*item.InstanceView.Statuses)[0]
			// first item in this field is the power status
			instancePowerStatus := (*item.InstanceView.Statuses)[1]

			if *instancePowerStatus.Code == "PowerState/running" {
				allocationCount++
			}

			if instanceProvisioningStatus.Time != nil {
				stateChangedAt = instanceProvisioningStatus.Time
			}
			// get the disk creation time to see when the instance was created initially
			createdAt = (*disk.Statuses)[0].Time
			// powerstate strings are of the format 'PowerState/running', 'PowerState/deallocated', etc.
			trimmedPowerState = strings.Split(*instancePowerStatus.Code, "/")[1]
		}
		formattedCreationTime := createdAt.ToTime().Format(time.RFC822)
		formattedStateChangeTime := stateChangedAt.ToTime().Format(time.RFC822)

		resultString := []string{key, trimmedPowerState, strconv.Itoa(allocationCount), formattedCreationTime, formattedStateChangeTime}
		resultData = append(resultData, resultString)
	}

	// post the results to a slack webhook
	postResultsToSlack(resultData)
}

func getAllVirtualMachines(ctx context.Context, client compute.VirtualMachinesClient) []compute.VirtualMachine {
	var virtualMachineList []compute.VirtualMachine
	vmPage, err := client.ListAll(ctx)
	if err != nil {
		log.Fatal("Failed to list Azure Virtual Machines", err)
	}
	for hasMorePages := true; hasMorePages; hasMorePages = (vmPage.NotDone()) {
		virtualMachineList = append(virtualMachineList, vmPage.Values()...)
	}
	return virtualMachineList
}

func getVirtualMachineWorker(ctx context.Context, client compute.VirtualMachinesClient, jobs <-chan string, results chan<- compute.VirtualMachine) {
	for virtualMachineName := range jobs {
		vm, err := client.Get(ctx, GetResourceGroupName(), virtualMachineName, compute.InstanceView)
		if err != nil {
			log.Println("Error fetching", virtualMachineName, err)
		}
		results <- vm
	}
}

func groupVirtualMachinesByExtendedInfo(ctx context.Context, client compute.VirtualMachinesClient, virtualMachines []compute.VirtualMachine) map[string][]compute.VirtualMachine {
	groupedVirtualMachines := make(map[string][]compute.VirtualMachine)
	fetchJobs := make(chan string, len(virtualMachines))
	fetchResults := make(chan compute.VirtualMachine, len(virtualMachines))

	for workers := 1; workers <= 5; workers++ {
		go getVirtualMachineWorker(ctx, client, fetchJobs, fetchResults)
	}
	for i := 0; i < len(virtualMachines); i++ {
		fetchJobs <- *virtualMachines[i].Name
	}
	close(fetchJobs)
	for j := 0; j < len(virtualMachines); j++ {
		result := <-fetchResults
		groupByTag := *result.Tags[GetGroupByPrefix()]
		groupedVirtualMachines[groupByTag] = append(groupedVirtualMachines[groupByTag], result)
	}
	return groupedVirtualMachines
}

func postResultsToSlack(resultTable [][]string) {
	buffer := new(bytes.Buffer)
	table := tablewriter.NewWriter(buffer)
	table.SetHeader([]string{"Key", "Power State", "Allocated", "Created", "Changed"})
	for _, value := range resultTable {
		table.Append(value)
	}
	table.Render()
	slackString := fmt.Sprintf("```%v```", buffer.String())
	payload := slack.Payload{
		Text:      slackString,
		Username:  "Azure",
		IconEmoji: ":robot_face:",
	}
	err := slack.Send(GetSlackWebhook(), "", payload)
	if len(err) > 0 {
		for errors := range err {
			log.Fatal("Error posting to Slack", errors)
		}
	}
}
