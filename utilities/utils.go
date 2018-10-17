package utilities

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	crmv1 "google.golang.org/api/cloudresourcemanager/v1beta1"
	crmv2 "google.golang.org/api/cloudresourcemanager/v2beta1"
	iam "google.golang.org/api/iam/v1"
)

//GetRoles gets the roles
func GetRoles() {
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	iamService, err := iam.New(client)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}
	response, err := iamService.Roles.List().Do()
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	for _, role := range response.Roles {
		log.Println(role.Name)
		log.Println(role.IncludedPermissions)
	}

}

//GetPermissions gets the permissions of a role
func GetPermissions(role string) {
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	iamService, err := iam.New(client)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	response, err := iamService.Roles.Get(role).Do()
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	for _, permission := range response.IncludedPermissions {
		// log.Println(role.Name)
		log.Println(permission)
		// log.Println(role.IncludedPermissions)
	}
}

//ConvertStructV2toV1 converts struct
func ConvertStructV2toV1(policy *crmv2.Policy) *crmv1.Policy {
	jblob, err := policy.MarshalJSON()
	if err != nil {
		log.Fatalln("DOH")
	}
	var result crmv1.Policy
	json.Unmarshal(jblob, &result)
	return &result
}

//FilterRolesByPermission filters a set of roles that contain a permission
func FilterRolesByPermission(roles []string, permission string) []string {
	results := make([]string, len(roles))
	client, err := google.DefaultClient(oauth2.NoContext, iam.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	iamService, err := iam.New(client)
	if err != nil {
		log.Fatalf("Couldn't create google client: %v", err)
	}

	for _, role := range roles {
		response, err := iamService.Roles.Get(role).Do()
		if err != nil {
			log.Fatalf("Couldn't create google client: %v", err)
		}
		if StringInSlice(permission, response.IncludedPermissions) {
			results = append(results, role)
		}

	}

	return results
}

//StringInSlice determines if a slice contains a string
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type conf struct {
	Core struct {
		Project string `yaml:"project"`
	}
}

//GetProjectID gets the project ID either from commandline or the default
func GetProjectID() string {
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
