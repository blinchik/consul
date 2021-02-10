package main

import (
	"fmt"
	"log"
	"os"

	secret "github.com/blinchik/aws/services/secretmanager"
	acl "github.com/blinchik/consul/acl"
)

var consulAddress string
var consulPort string
var consulRootPath string

func main() {

	consulAddress = os.Args[1]
	consulPort = os.Args[2]
	consulRootPath = os.Args[3]

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	output := acl.BootstrapACL(consulAddress, consulRootPath, consulPort)

	if output != nil {
		secretName := fmt.Sprintf("brain-%s", output.Policies[0].Name)
		secret.CreateSecret(secretName, output.SecretID, output.Description)

	}

}
