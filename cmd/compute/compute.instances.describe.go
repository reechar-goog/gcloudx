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
	"strings"

	"github.com/ghodss/yaml"
	gcloudx "github.com/reechar-goog/gcloudx/cmd"
	utils "github.com/reechar-goog/gcloudx/utilities"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	iam "google.golang.org/api/iam/v1"
)

var computeInstancesDescribeCmd = &cobra.Command{
	Use:              "describe",
	TraverseChildren: false,
	Short:            "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		doDescribe()
	},
}

func doDescribe() {
	projectID := utils.GetProjectIDFromGcloud()
	instanceID := getInstanceID()
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	computeService, err := compute.New(client)
	if err != nil {
		log.Fatalf("Couldn't create Compute Service: %v", err)
	}
	if *gcloudx.Zone != "" {

	}
	// zones, err := computeService.Zones.List(projectID).Do()
	zone := "us-east1-b"
	instance, err := computeService.Instances.Get(projectID, zone, instanceID).Do()
	if err != nil {
		log.Fatalf("Couldn't get instance: %v\n", err)
	}
	diskToImage := make(map[string]string)
	for _, disk := range instance.Disks {
		indexOfDiskName := strings.LastIndex(disk.Source, "/") + 1
		diskName := disk.Source[indexOfDiskName:]
		diskInfo, _ := computeService.Disks.Get(projectID, zone, diskName).Do()
		diskToImage[disk.Source] = diskInfo.SourceImage
	}

	json, err := instance.MarshalJSON()
	yaml, err := yaml.JSONToYAML(json)
	stringYaml := string(yaml)
	for disk, image := range diskToImage {
		replaceString := disk + "\n  sourceImage: " + image
		stringYaml = strings.Replace(stringYaml, disk, replaceString, 1)
	}
	fmt.Printf(stringYaml)
}

func init() {
	computeInstancesCmd.AddCommand(computeInstancesDescribeCmd)
}

func getInstanceID() string {
	lastArg := pflag.Args()[len(pflag.Args())-1]
	if lastArg != "describe" {
		return lastArg
	}
	return ""
}
