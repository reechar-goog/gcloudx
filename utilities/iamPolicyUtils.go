package utilities

import (
	"log"
	"strings"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1beta1"
	crmv2 "google.golang.org/api/cloudresourcemanager/v2beta1"
	iam "google.golang.org/api/iam/v1"
)

//GetFullyInheritedPolicy get IAM Policy with inheritance
func GetFullyInheritedPolicy(projectID string) *crmv1.Policy {
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

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
	projectPolicy, err := projectsService.Projects.GetIamPolicy(projectID, gipr).Do()
	resultsPolicy = projectPolicy

	var parentType = proj.Parent.Type
	if parentType == "folder" {
		var parentID = "folders/" + proj.Parent.Id
		folder, err := foldersService.Folders.Get(parentID).Do()
		folderPolicy, err := foldersService.Folders.GetIamPolicy(parentID, new(crmv2.GetIamPolicyRequest)).Do()

		if err != nil {
			log.Fatalf("Unable to get folder policy: %v", err)
		}
		resultsPolicy = MergePolicy(resultsPolicy, ConvertStructV2toV1(folderPolicy))
		var currentParent = folder.Parent

		for strings.HasPrefix(currentParent, "folder") {
			var nextParent, err = foldersService.Folders.Get(currentParent).Do()
			if err != nil {
				log.Fatalf("Unable to get folder policy: %v", err)
			}
			folderPolicy, err := foldersService.Folders.GetIamPolicy(nextParent.Name, new(crmv2.GetIamPolicyRequest)).Do()
			resultsPolicy = MergePolicy(resultsPolicy, ConvertStructV2toV1(folderPolicy))
			currentParent = nextParent.Parent
		}
	}

	var parentID = "organizations/" + proj.Parent.Id
	orgPolicy, err := projectsService.Organizations.GetIamPolicy(parentID, gipr).Do()
	if err != nil {
		log.Fatalf("Unable to get folder policy: %v", err)
	}

	resultsPolicy = MergePolicy(resultsPolicy, orgPolicy)
	return resultsPolicy
}

//MergePolicy merges IAM policies
func MergePolicy(policy1, policy2 *crmv1.Policy) *crmv1.Policy {
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
			mergedBinding := new(crmv1.Binding)
			mergedBinding.Role = binding.Role
			polset := mapset.NewSet()
			for _, v := range binding.Members {
				polset.Add(v)
			}
			for _, v := range roleMap[binding.Role].Members {
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

//FilterPolicyByRole filters policy
func FilterPolicyByRole(policy *crmv1.Policy, roleFilter []string) *crmv1.Policy {
	filteredPolicy := new(crmv1.Policy)
	filteredBindings := make([]*crmv1.Binding, 0)

	for _, binding := range policy.Bindings {
		if StringInSlice(binding.Role, roleFilter) {
			filteredBindings = append(filteredBindings, binding)
		}
	}

	filteredPolicy.Bindings = filteredBindings
	return filteredPolicy
}

//GetRolesFromPolicy gets all roles defined in a policy
func GetRolesFromPolicy(policy *crmv1.Policy) []string {
	roles := make([]string, 0)
	for _, binding := range policy.Bindings {
		roles = append(roles, binding.Role)
	}
	return roles
}
