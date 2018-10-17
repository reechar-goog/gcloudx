package main

import (
	"github.com/reechar-goog/gcloudx/cmd"

	"github.com/spf13/pflag"
)

func main() {
	pflag.Parse()
	// util.GetRoles()
	// util.GetPermissions("roles/iam.serviceAccountKeyAdmin")
	cmd.Execute()
}
