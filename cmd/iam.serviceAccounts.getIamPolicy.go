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

package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	yaml "github.com/ghodss/yaml"
	utils "github.com/reechar-goog/gcloudx/utilities"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	iam "google.golang.org/api/iam/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var saiampolicyCmd = &cobra.Command{
	Use:              "get-iam-policy",
	TraverseChildren: false,
	Short:            "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		doSaIam()
	},
}

func doSaIam() {
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	serviceAccountID := getServiceAccount()
	iamService, err := iam.New(client)
	if err != nil {
		log.Fatalf("Couldn't create IAM Service: %v", err)
	}
	sa, err := iamService.Projects.ServiceAccounts.Get("projects/-/serviceAccounts/" + serviceAccountID).Do()
	if err != nil {
		log.Fatalf("Couldn't get service account: %v", err)
	}
	projectID := sa.ProjectId

	serviceAccountPolicy, err := iamService.Projects.ServiceAccounts.GetIamPolicy("projects/" + projectID + "/serviceAccounts/" + serviceAccountID).Do()

	projectPolicy := utils.GetFullyInheritedPolicy(projectID)
	results := utils.MergePolicy(projectPolicy, utils.ConvertStructIamtoV1(serviceAccountPolicy))

	if *roles != "" {
		roleFilter := strings.Split(*roles, ",")
		results = utils.FilterPolicyByRole(results, roleFilter)
	}

	if *permission != "" {
		roles := getRoles(results)
		roleFilter := utils.FilterRolesByPermission(roles, *permission)
		results = utils.FilterPolicyByRole(results, roleFilter)
	}

	resultsJSON, err := results.MarshalJSON()
	if err != nil {
		log.Fatalf("Could not parse json: %v", err)
	}
	resultsYaml, err := yaml.JSONToYAML(resultsJSON)
	if err != nil {
		log.Fatalf("Could not prase yaml: %v", err)
	}
	fmt.Println(string(resultsYaml))

}

func getServiceAccount() string {
	lastArg := pflag.Args()[len(pflag.Args())-1]
	if lastArg == "get-iam-policy" {
		log.Fatalln("ERROR: usage: gcloudx iam service-accounts <service-account-id>")
	}
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(lastArg) {
		log.Fatalln("ERROR: Service Account ID format must be a valid e-mail")
	}
	return lastArg
}

func init() {
	saCmd.AddCommand(saiampolicyCmd)
}
