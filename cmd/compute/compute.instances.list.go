// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	utils "github.com/reechar-goog/gcloudx/utilities"
	"github.com/spf13/cobra"
)

var computeInstancesListCmd = &cobra.Command{
	Use:              "list",
	TraverseChildren: false,
	Short:            "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		doComputeList()
	},
}

func doComputeList() {
	fmt.Println("Listing computes")
	// client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	// if err != nil {
	// 	log.Fatalf("Couldn't create google client: %v", err)
	// }

	// computeService, err := compute.New(client)
	// if err != nil {
	// 	log.Fatalf("Couldn't create Compute Service: %v", err)
	// }
	projectID := utils.GetProjectIDFromGcloud()
	// // zone := "my-zone"
	// zones, err := computeService.Zones.List(projectID).Do()
	// for _, zone := range zones.Items {
	// 	fmt.Println(zone.Name)
	// }

	cmd := exec.Command("gcloud", "compute", "instances", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}
	stringout := string(out)
	lines := strings.Split(stringout, "\n")
	lines[0] = lines[0] + "      CONSOLE_LINK"
	//https://console.cloud.google.com/compute/instancesDetail/zones/us-east/instances/asd?project=reechar-dm-preempt
	// fmt.Printf(string(out))
	for i, line := range lines {
		if i == 0 {
			continue
		}
		lines[i] = lines[i] + "https://console.cloud.google.com/compute/instancesDetail/?project" + projectID
		fmt.Printf(strconv.Itoa(i) + ": " + line + "\n")
	}

}

func init() {
	computeInstancesCmd.AddCommand(computeInstancesListCmd)
}
