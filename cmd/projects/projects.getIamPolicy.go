package cmd

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	yaml "github.com/ghodss/yaml"
	gcloudx "github.com/reechar-goog/gcloudx/cmd"
	utils "github.com/reechar-goog/gcloudx/utilities"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1beta1"
)

var getIamPolicyCmd = &cobra.Command{
	Use:              "get-iam-policy",
	TraverseChildren: false,
	Short:            "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		doProjectIam()
	},
}

func doProjectIam() {
	var projectID = getProjectID()
	results := utils.GetFullyInheritedPolicy(projectID)

	if *gcloudx.Roles != "" {
		roleFilter := strings.Split(*gcloudx.Roles, ",")
		results = utils.FilterPolicyByRole(results, roleFilter)
	}

	if *gcloudx.Permission != "" {
		roles := utils.GetRolesFromPolicy(results)
		roleFilter := utils.FilterRolesByPermission(roles, *gcloudx.Permission)
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

func getRoles(policy *crmv1.Policy) []string {
	roles := make([]string, 0)
	for _, binding := range policy.Bindings {
		roles = append(roles, binding.Role)
	}
	return roles
}

func init() {
	projectsCmd.AddCommand(getIamPolicyCmd)
}

type conf struct {
	Core struct {
		Project string `yaml:"project"`
	}
}

func getProjectID() string {
	lastArg := pflag.Args()[len(pflag.Args())-1]
	if lastArg != "get-iam-policy" {
		return lastArg
	}
	cmd := exec.Command("gcloud", "config", "list", "--format", "yaml")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	c := conf{}
	err = yaml.Unmarshal(out, &c)
	if err != nil {
		log.Fatal(err)
	}
	return c.Core.Project
}
