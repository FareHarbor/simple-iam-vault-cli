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
func VaultLogin(role string, loginData map[string]interface{}, onlyToken bool) {
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
	if onlyToken {
		fmt.Println(parseResponse(string(body)))
	} else {
		fmt.Println(jsonPrettyPrint(string(body)))
	}
}

type HttpResponse struct {
	Auth *HttpAuth `json:"auth"`
}

type HttpAuth struct {
	ClientToken string `json:"client_token"`
}

func parseResponse(in string) string {
	var httpResponse HttpResponse
	err := json.Unmarshal([]byte(in), &httpResponse)
	if err != nil {
		panic(err)
	}
	return httpResponse.Auth.ClientToken
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
	var region string
	var role string
	var headerHost string
	var onlyToken bool
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "region",
				Usage:       "AWS region we are using (e.g. us-west-1)",
				Destination: &region,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "role",
				Usage:       "Vault role we are using",
				Destination: &role,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "host",
				Usage:       "Host for the  X-Vault-AWS-IAM-Server-ID header",
				Destination: &headerHost,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "only-token",
				Usage:       "If present, return the vault token only",
				Destination: &onlyToken,
			},
		},
		Action: func(c *cli.Context) error {
			loginData, _ := GenerateLoginData(region, headerHost)
			VaultLogin(role, loginData, onlyToken)
			return nil
		},
		Name:  "simple-iam-vault-cli - a 'simple' IAM vault login CLI",
		Usage: "VAULT_ADDR=[vault url] ./simple-iam-vault-cli --region [AWS region] --role [Vault role] --host [Host for Server-ID header] --only-token (optional)",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
