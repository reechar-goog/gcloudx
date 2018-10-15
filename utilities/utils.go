package utilities

import (
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
