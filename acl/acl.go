package acl

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var consulAddress string
var consulPort string
var consulRootPath string
var prefix string
var clientKeyFile string
var clientCertFile string
var caChainFile string

func main() {

	consulAddress = os.Args[1]
	consulPort = os.Args[2]
	consulRootPath = os.Args[3]
	prefix = os.Args[4]
	clientKeyFile = os.Args[5]
	clientCertFile = os.Args[6]
	caChainFile = os.Args[7]

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	output := BootstrapACL(consulAddress, consulRootPath, consulPort, clientCertFile, clientKeyFile, caChainFile)

	fmt.Print(output)

}

//BootstrapACLResponse the response receviced from bootstraping consul server
type BootstrapACLResponse struct {
	ID          string `json:"ID"`
	AccessorID  string `json:"AccessorID"`
	SecretID    string `json:"SecretID"`
	Description string `json:"Description"`
	Policies    []struct {
		ID   string `json:"ID"`
		Name string `json:"Name"`
	} `json:"Policies"`
	Local       bool   `json:"Local"`
	CreateTime  string `json:"CreateTime"`
	Hash        string `json:"Hash"`
	CreateIndex int    `json:"CreateIndex"`
	ModifyIndex int    `json:"ModifyIndex"`
}

//BootstrapACL This endpoint does a special one-time bootstrap of the ACL system, making the first management token if the acl.tokens.master configuration entry is not
//specified in the Consul server configuration and if the cluster has not been bootstrapped previously.
func BootstrapACL(consulAddress, consulRootPath, consulPort, clientCertFile, clientKeyFile, caChainFile string) *BootstrapACLResponse {

	var output BootstrapACLResponse

	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)

	if err != nil {
		log.Fatal(err)
	}

	caCert, err := ioutil.ReadFile(caChainFile)

	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}

	client := http.Client{Transport: t, Timeout: 15 * time.Second}

	req, err := http.NewRequest("PUT", fmt.Sprintf("https://%s:%s/%s/acl/bootstrap", consulAddress, consulPort, consulRootPath), nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bodyBytes, &output)
	if err != nil {

		fmt.Println(string(bodyBytes))

		if strings.Contains(string(bodyBytes), "ACL bootstrap no longer allowed") {

			return nil

		} else {

			log.Fatal(err)

		}
	}

	return &output

}

//UpdateACLToken existing ACL token
func UpdateACLToken(consulAddress, consulRootPath, consulPort, token, consulToken string) {

	payload := fmt.Sprintf(` {"token": "%s" }`, token)

	body := strings.NewReader(payload)

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%s/%s/agent/token/agent", consulAddress, consulPort, consulRootPath), body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", consulToken))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(bodyBytes))

}
