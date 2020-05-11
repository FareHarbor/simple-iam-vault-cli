package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/urfave/cli"
)

// GenerateLoginData is a stripped down version of Vault CLI's code
// https://github.com/hashicorp/vault/blob/master/builtin/credential/aws/cli.go#L38
// It assumes that we are using the X-Vault-AWS-IAM-Server-ID header.
func GenerateLoginData(configuredRegion string, configuredHost string) (map[string]interface{}, error) {

	loginData := make(map[string]interface{})
	s, _ := session.NewSession(&aws.Config{
		Region: aws.String(configuredRegion),
	})
	svc := sts.New(s)

	var params *sts.GetCallerIdentityInput
	stsRequest, _ := svc.GetCallerIdentityRequest(params)
	stsRequest.HTTPRequest.Header.Add("X-Vault-AWS-IAM-Server-ID", configuredHost)
	stsRequest.Sign()

	headersJSON, err := json.Marshal(stsRequest.HTTPRequest.Header)
	if err != nil {
		return nil, err
	}

	requestBody, err := ioutil.ReadAll(stsRequest.HTTPRequest.Body)
	if err != nil {
		return nil, err
	}

	loginData["iam_http_request_method"] = stsRequest.HTTPRequest.Method
	loginData["iam_request_url"] = base64.StdEncoding.EncodeToString([]byte(stsRequest.HTTPRequest.URL.String()))
	loginData["iam_request_headers"] = base64.StdEncoding.EncodeToString(headersJSON)
	loginData["iam_request_body"] = base64.StdEncoding.EncodeToString(requestBody)

	return loginData, nil
}

// VaultLogin is the function that takes the login data and sends it to Vault.
func VaultLogin(role string, loginData map[string]interface{}) {
	var vaultAddr = os.Getenv("VAULT_ADDR")
	//TODO: Make configurable
	awsAuthPath := "auth/aws"
	path := vaultAddr + "/v1/" + awsAuthPath + "/login"
	loginData["role"] = role

	jsonStr, _ := json.Marshal(loginData)
	request, _ := http.NewRequest("POST", path, bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(jsonPrettyPrint(string(body)))
}

// jsonPrettyPrint borrowed from https://stackoverflow.com/a/36544455/5981682
func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "region, r",
				Usage:    "AWS region we are using (e.g. us-west-1)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "role, ro",
				Usage:    "Vault role we are using",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "host, h",
				Usage:    "Host for the  X-Vault-AWS-IAM-Server-ID header",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "url, u",
				Usage:    "Vault URL (e.g. https://vault.example.com:8200)",
				EnvVars:  "VAULT_ADDR",
				Required: true,
			},
		},

		Name:  "simple-iam-vault-cli",
		Usage: "Relatively simple AWS IAM login to Hashicorp Vault",
		Action: func(c *cli.Context) error {
			region := os.Args[1]
			role := os.Args[2]
			host := os.Args[3]
			loginData, _ := GenerateLoginData(region, host)
			VaultLogin(role, loginData)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
