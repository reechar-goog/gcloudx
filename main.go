package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
)

func doNormal() {
	cmd := exec.Command("gcloud", flag.Args()...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s\n", string(out))
}
func doSpecial() {
	fmt.Printf("trolol")
}

func main() {
	fmt.Println("https://console.cloud.google.com/compute/instancesDetail/zones/us-central1-c/instances/certbot?project=reechar-kubernetes")
	flag.Parse()
	specialCommands := make(map[string]func())
	specialCommands["special action"] = doSpecial
	specialCommands["projects get-iam-policy"] = doProjectIam

	commandKey := strings.Join(flag.Args(), " ")
	if val, ok := specialCommands[commandKey]; ok {
		val()
	} else {
		doNormal()
	}

	// err := cmd.Run()
	// stdOut, err := cmd.StdoutPipe()

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// errorr := cmd.Wait()
	// if errorr != nil {
	// 	log.Fatal(errorr)
	// }
	// fmt.Printf("%s\n", stdOut)

}
