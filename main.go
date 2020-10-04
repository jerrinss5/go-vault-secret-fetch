package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	vault "github.com/hashicorp/vault/api"
)

// VaultAuthHeaderName - theft protection
const VaultAuthHeaderName = "X-Vault-AWS-IAM-Server-ID"

var (
	vaultAddr         string
	vaultAuthProvider string
	vaultAuthRole     string
	vaultAuthHeader   string
	vaultClient       *vault.Client
	token             string
	tokenIsRenewable  bool
	tokenExpiration   time.Time     // actual expiration
	tokenTTL          time.Duration // lifetime of the auth token received
	expirationWindow  time.Duration // time to allow to process a token renewal
	renewalWindow     time.Duration // time before expiration when token should be actively renewed
)

func init() {
	//  vaultAddr = os.Getenv("VAULT_ADDR")
	// vaultAuthProvider = os.Getenv("VAULT_AUTH_PROVIDER")
	// vaultAuthRole = os.Getenv("VAULT_AUTH_ROLE")
	// vaultAuthHeader = os.Getenv("VAULT_AUTH_HEADER")
	vaultAddr = "http://127.0.0.1:8200"
	vaultAuthProvider = "aws"
	vaultAuthRole = "example-role-name"
	vaultAuthHeader = "vault.example.com"

	vaultClient, _ = vault.NewClient(&vault.Config{Address: vaultAddr})
}

func parseToken(resp *vault.Secret) error {
	var err error
	if token, err = resp.TokenID(); err != nil {
		return err
	}

	if tokenIsRenewable, err = resp.TokenIsRenewable(); err != nil {
		return err
	}

	if tokenTTL, err = resp.TokenTTL(); err != nil {
		return err
	}
	tokenExpiration = time.Now().Add(tokenTTL)

	vaultClient.SetToken(token)
	return nil
}

func awsLogin() error {

	if vaultAddr == "" || vaultAuthProvider == "" || vaultAuthRole == "" {
		return fmt.Errorf("you must set the VAULT_ADDR, VAULT_AUTH_PROVIDER, and VAULT_AUTH_ROLE environment variables")
	}

	stsSvc := sts.New(session.New())
	fmt.Println("Calling AWS STS Service to get the Identity")
	req, _ := stsSvc.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})

	if vaultAuthHeader != "" {
		// if supplied, and then sign the request including that header
		req.HTTPRequest.Header.Add(VaultAuthHeaderName, vaultAuthHeader)
	}
	req.Sign()

	headers, err := json.Marshal(req.HTTPRequest.Header)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(req.HTTPRequest.Body)
	if err != nil {
		return err
	}

	d := make(map[string]interface{})
	d["iam_http_request_method"] = req.HTTPRequest.Method
	d["iam_request_url"] = base64.StdEncoding.EncodeToString([]byte(req.HTTPRequest.URL.String()))
	d["iam_request_headers"] = base64.StdEncoding.EncodeToString(headers)
	d["iam_request_body"] = base64.StdEncoding.EncodeToString(body)
	d["role"] = vaultAuthRole

	fmt.Println("Calling Vault with identity from STS service")
	resp, err := vaultClient.Logical().Write(fmt.Sprintf("auth/%s/login", vaultAuthProvider), d)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("Got no response from the %s authentication provider", vaultAuthProvider)
	}
	fmt.Println("Got Vault Token")
	return parseToken(resp)
}

func getSecret() {
	secret, err := vaultClient.Logical().Read("secret/data/data/foo")
	if err != nil {
		panic(err)
	}
	b, _ := json.Marshal(secret.Data)
	fmt.Println(string(b))
}

func main() {
	err := awsLogin()
	if err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}
	// if no error get secret
	getSecret()
}
