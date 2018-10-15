package cmd

import (
	"fmt"
	utils "gcloudx/utilities"
	"log"
	"net/http"
	"os/exec"
	"strings"

	mapset "github.com/deckarep/golang-set"
	yaml "github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1beta1"
	crmv2 "google.golang.org/api/cloudresourcemanager/v2beta1"
	iam "google.golang.org/api/iam/v1"
)

// getIamPolicyCmd represents the getIamPolicy command
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

func init() {
	projectsCmd.AddCommand(getIamPolicyCmd)
}

type conf struct {
	Core struct {
		// Account string `yaml:"account"`
		// DUR     string `yaml:"disable_usage_reporting"`
		Project string `yaml:"project"`
	}
}

func getRoles(policy *crmv1.Policy) []string {
	roles := make([]string, 0)
	for _, binding := range policy.Bindings {
		roles = append(roles, binding.Role)
	}
	return roles
}

func doProjectIam() {

	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	var projectID = getProjectID()
	results := getHeirachy(client, projectID)

	if *roles != "" {
		roleFilter := strings.Split(*roles, ",")
		results = filterPolicy(results, roleFilter)
	}

	if *permission != "" {
		roles := getRoles(results)
		roleFilter := utils.FilterRolesByPermission(roles, *permission)
		results = filterPolicy(results, roleFilter)
	}

	resultsJSON, err := results.MarshalJSON()
	resultsYaml, err := yaml.JSONToYAML(resultsJSON)
	fmt.Println(string(resultsYaml))
}

func getProjectID() string {
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
func filterPolicy(policy *crmv1.Policy, roleFilter []string) *crmv1.Policy {
	filteredPolicy := new(crmv1.Policy)
	filteredBindings := make([]*crmv1.Binding, 0)

	for _, binding := range policy.Bindings {
		if utils.StringInSlice(binding.Role, roleFilter) {
			filteredBindings = append(filteredBindings, binding)
		}
	}

	filteredPolicy.Bindings = filteredBindings

	return filteredPolicy
}

func getHeirachy(client *http.Client, projectID string) *crmv1.Policy {

	var resultsPolicy *crmv1.Policy
	projectsService, err := crmv1.New(client)
	if err != nil {
		log.Fatalf("Unable to create Project service: %v", err)
	}
	foldersService, err := crmv2.New(client)
	if err != nil {
		log.Fatalf("Unable to create Resource Manager service: %v", err)
	}

	proj, err := projectsService.Projects.Get(projectID).Do()
	if err != nil {
		log.Fatalf("Unable to get project: %v", err)
	}
	var gipr = new(crmv1.GetIamPolicyRequest)
	policy, err := projectsService.Projects.GetIamPolicy(projectID, gipr).Do()
	// policyJSON, err := policy.MarshalJSON()
	// policyYAML, err := yaml.JSONToYAML(policyJSON)
	// _ = policyYAML
	// log.Println("Project: ")
	// log.Println(string(policyYAML))

	var ptype = proj.Parent.Type
	if ptype == "folder" {
		var parentID = "folders/" + proj.Parent.Id
		folder, err := foldersService.Folders.Get(parentID).Do()
		fpo, err := foldersService.Folders.GetIamPolicy(parentID, new(crmv2.GetIamPolicyRequest)).Do()
		_ = fpo
		// log.Println("FODLER POLICY")
		// fpojson, err := fpo.MarshalJSON()
		// log.Println(string(fpojson))
		if err != nil {
			log.Fatalf("can't get")
		}
		// results = append(results, parentID)
		var currentParent = folder.Parent
		// fik
		for strings.HasPrefix(currentParent, "folder") {
			var nextParent, err = foldersService.Folders.Get(currentParent).Do()
			if err != nil {
				log.Fatalf("can't get")
			}
			// results = append(results, currentParent)
			currentParent = nextParent.Parent
		}
		// results = append(results, currentParent)
		policy2, err := projectsService.Organizations.GetIamPolicy(currentParent, gipr).Do()
		if err != nil {
			log.Fatalln("COULDn'T GET POLICYs")
		}
		policyJSON2, err := policy2.MarshalJSON()
		// log.Println("ORG POLICY:")
		log.Println(string(policyJSON2))

	} else if ptype == "organization" {
		var parentID = "organizations/" + proj.Parent.Id
		// results = append(results, parentID)
		policy2, err := projectsService.Organizations.GetIamPolicy(parentID, gipr).Do()
		if err != nil {
			log.Fatalln("COULDn'T GET POLICYs")
		}
		// crmv1.Policy.Bindings[0].
		// policyJSON2, err := policy2.MarshalJSON()
		// policyYAML2, err := yaml.JSONToYAML(policyJSON2)

		// log.Println("ORG POLICY:")
		// log.Println(string(policyYAML2))
		mergedPolicy := mergePolicy(policy, policy2)
		resultsPolicy = mergedPolicy

		// policyJSON3, err := mergedPolicy.MarshalJSON()
		// policyYAML3, err := yaml.JSONToYAML(policyJSON3)
		// log.Println("MERGED")
		// fmt.Println(string(policyYAML3))
	}

	return resultsPolicy
}

func mergePolicy(policy1, policy2 *crmv1.Policy) *crmv1.Policy {
	//  var mergedPolicy crmv1.olicy
	roleMap := make(map[string]crmv1.Binding)

	policy1Bindings := mapset.NewSet()
	for _, binding := range policy1.Bindings {
		policy1Bindings.Add(binding.Role)
	}

	policy2Bindings := mapset.NewSet()
	for _, binding := range policy2.Bindings {
		policy2Bindings.Add(binding.Role)
	}

	policyIntersection := policy1Bindings.Intersect(policy2Bindings)

	mergedPolicy := new(crmv1.Policy)
	mergedPolicyBindings := make([]*crmv1.Binding, 0)

	for _, binding := range policy1.Bindings {
		if !policyIntersection.Contains(binding.Role) {
			mergedPolicyBindings = append(mergedPolicyBindings, binding)
		} else {
			roleMap[binding.Role] = *binding
		}
	}

	for _, binding := range policy2.Bindings {
		if !policyIntersection.Contains(binding.Role) {
			mergedPolicyBindings = append(mergedPolicyBindings, binding)
		} else {
			// log.Printf("merging role: %s", binding.Role)
			mergedBinding := new(crmv1.Binding)
			mergedBinding.Role = binding.Role
			polset := mapset.NewSet()
			for _, v := range binding.Members {
				// log.Printf("adding %s from pol1", v)
				polset.Add(v)
			}
			for _, v := range roleMap[binding.Role].Members {
				// log.Printf("adding %s from pol2", v)
				polset.Add(v)
			}
			slicer := polset.ToSlice()
			stringers := make([]string, len(slicer))
			for i, v := range slicer {
				stringers[i] = v.(string)
			}
			mergedBinding.Members = stringers
			mergedPolicyBindings = append(mergedPolicyBindings, mergedBinding)
		}
	}

	mergedPolicy.Bindings = mergedPolicyBindings
	return mergedPolicy
}
