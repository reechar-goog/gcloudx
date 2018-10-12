package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	yaml "github.com/ghodss/yaml"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1beta1"
	crmv2 "google.golang.org/api/cloudresourcemanager/v2beta1"
)

type conf struct {
	Core struct {
		// Account string `yaml:"account"`
		// DUR     string `yaml:"disable_usage_reporting"`
		Project string `yaml:"project"`
	}
}

type IAMPolicy struct {
}

func doProjectIam() {
	client, err := google.DefaultClient(oauth2.NoContext)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}
	var projectID = getProjectID()
	_ = projectID
	// results := getHeirachy(client, projectID)
	// log.Println(results)

	fmt.Println(getProjectID())

	// svc, err := iam.New(client)
	// pol, err := svc.Projects.ServiceAccounts.GetIamPolicy("projects/p-reech-api-tf9/serviceAccounts/p-reech-api-tf9@appspot.gserviceaccount.com").Do()
	// pjson, err := pol.MarshalJSON()
	// log.Println("SA JSON")
	// log.Println(string(pjson))
}

func getProjectID() string {
	// log.Printf("here i am")
	// cmd := exec.Command("gcloud", "config", "list", "--format", "yaml")
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s\n", string(out))
	// c := conf{}
	// err = yaml.Unmarshal(out, &c)
	// // fmt.Printf("--- m:\n%v\n\n", c)
	// // fmt.Printf("yo" + c.Core.Project)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("%s", out)

	return strings.TrimSpace(string(out))
}

func getHeirachy(client *http.Client, projectID string) []string {
	var results []string
	results = append(results, projectID)
	projectsService, err := crmv1.New(client)
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
	policyJSON, err := policy.MarshalJSON()
	policyYAML, err := yaml.JSONToYAML(policyJSON)
	log.Println(string(policyYAML))

	var ptype = proj.Parent.Type
	if ptype == "folder" {
		var parentID = "folders/" + proj.Parent.Id
		folder, err := foldersService.Folders.Get(parentID).Do()
		fpo, err := foldersService.Folders.GetIamPolicy(parentID, new(crmv2.GetIamPolicyRequest)).Do()
		log.Println("FODLER POLICY")
		fpojson, err := fpo.MarshalJSON()
		log.Println(string(fpojson))
		if err != nil {
			log.Fatalf("can't get")
		}
		results = append(results, parentID)
		var currentParent = folder.Parent
		// fik
		for strings.HasPrefix(currentParent, "folder") {
			var nextParent, err = foldersService.Folders.Get(currentParent).Do()
			if err != nil {
				log.Fatalf("can't get")
			}
			results = append(results, currentParent)
			currentParent = nextParent.Parent
		}
		results = append(results, currentParent)
		policy2, err := projectsService.Organizations.GetIamPolicy(currentParent, gipr).Do()
		if err != nil {
			log.Fatalln("COULDn'T GET POLICYs")
		}
		policyJSON2, err := policy2.MarshalJSON()
		log.Println("ORG POLICY:")
		log.Println(string(policyJSON2))

	} else if ptype == "organization" {
		var parentID = "organizations/" + proj.Parent.Id
		results = append(results, parentID)
		policy2, err := projectsService.Organizations.GetIamPolicy(parentID, gipr).Do()
		if err != nil {
			log.Fatalln("COULDn'T GET POLICYs")
		}
		// crmv1.Policy.Bindings[0].
		policyJSON2, err := policy2.MarshalJSON()
		policyYAML2, err := yaml.JSONToYAML(policyJSON2)

		log.Println("ORG POLICY:")
		log.Println(string(policyYAML2))
		pol := mergePolicy(policy, policy2)
		_ = pol
		policyJSON3, err := pol.MarshalJSON()
		policyYAML3, err := yaml.JSONToYAML(policyJSON3)
		log.Println("MERGED")
		log.Println(string(policyYAML3))
	}

	return results
}

func mergePolicy(policy1, policy2 *crmv1.Policy) *crmv1.Policy {
	//  var mergedPolicy crmv1.olicy
	roleMap := make(map[string]crmv1.Binding)

	pol1bindings := mapset.NewSet()
	for _, binding := range policy1.Bindings {
		pol1bindings.Add(binding.Role)
	}
	fmt.Println("Policy 1 roles:")
	fmt.Println(pol1bindings)

	pol2bindings := mapset.NewSet()
	for _, binding := range policy2.Bindings {
		pol2bindings.Add(binding.Role)
	}
	fmt.Println("Policy 2 roles:")
	fmt.Println(pol2bindings)

	polinter := pol1bindings.Intersect(pol2bindings)
	fmt.Println("Intersection")
	fmt.Println(polinter)

	for _, binding := range policy1.Bindings {
		roleMap[binding.Role] = *binding
	}

	// pol1bindings.

	// log.Println("whatsup")
	// // log.Println(roleMap)
	// for k, v := range roleMap {
	// 	bindingTest := new(crmv1.Binding)
	// 	bindingTest.Role = k
	// 	bindingTestMembers := append(v.Members, "TESTEROOO")
	// 	bindingTest.Members = bindingTestMembers
	// 	roleMap[k] = (*bindingTest)
	// }
	// for k, v := range roleMap {
	// 	fmt.Printf("key[%s] value[%s]\n", k, v.Members)
	// }

	mergedPolicy := new(crmv1.Policy)
	mergedPolicyBindings := make([]*crmv1.Binding, 0)

	for _, binding := range policy1.Bindings {
		if !polinter.Contains(binding.Role) {
			mergedPolicyBindings = append(mergedPolicyBindings, binding)
		}
	}

	for _, binding := range policy2.Bindings {
		if !polinter.Contains(binding.Role) {
			mergedPolicyBindings = append(mergedPolicyBindings, binding)
		} else {
			mergedBinding := new(crmv1.Binding)
			mergedBinding.Role = binding.Role
			// s0 := []interface{} binding.Members
			polset := mapset.NewSet()
			for _, v := range binding.Members {
				polset.Add(v)
			}
			for _, v := range binding.Members {
				polset.Add(v)
			}
			// pol1Set := mapset.NewSetFromSlice(binding.Members.([]interface{}))
			// pol2Set := mapset.NewSet(roleMap[binding.Role].Members)
			// pol3 := pol1Set.Union(pol2Set)
			slicer := polset.ToSlice()
			stringers := make([]string, len(slicer))
			for i, v := range slicer {
				stringers[i] = v.(string)
			}
			mergedBinding.Members = stringers
			// mergedBinding.Members = (pol1Set.Union(pol2Set)).ToSlice()
			mergedPolicyBindings = append(mergedPolicyBindings, mergedBinding)
		}
	}

	// newBinding := append(policy1.Bindings, bindingTest)
	mergedPolicy.Bindings = mergedPolicyBindings
	return mergedPolicy
}
