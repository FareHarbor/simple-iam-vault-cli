# simple-iam-vault-cli

Minimalist binary to authenticate to vault using IAM Instance Role.

Expects vault base url to be set using VAULT_ADDR env var

Usage: `./simple-iam-vault-cli REGION VAULT_ROLE`

## Building

**Local**: `go mod download && go build`

## Special thanks...

...to stormbeta for doing the work [here](https://github.com/stormbeta/snippets/tree/master/golang/vault-iam-auth).