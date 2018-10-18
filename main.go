package main

import (
	"github.com/reechar-goog/gcloudx/cmd"
	compute "github.com/reechar-goog/gcloudx/cmd/compute"
	iam "github.com/reechar-goog/gcloudx/cmd/iam"
	projects "github.com/reechar-goog/gcloudx/cmd/projects"

	"github.com/spf13/pflag"
)

func main() {
	compute.Load()
	projects.Load()
	iam.Load()
	pflag.Parse()
	// util.GetRoles()
	// util.GetPermissions("roles/iam.serviceAccountKeyAdmin")
	cmd.Execute()
}
