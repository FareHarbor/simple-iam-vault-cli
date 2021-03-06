# simple-iam-vault-cli

![Build Tag Release](https://github.com/bermannoah/simple-iam-vault-cli/workflows/Build%20Tag%20Release/badge.svg)

A 'simple' binary to authenticate to vault using IAM Instance Role.

Expects vault base url to be set using VAULT_ADDR envvar, or you can pass when invoking.

Usage: `VAULT_ADDR=[vault url] ./simple-iam-vault-cli --region [AWS region] --role [Vault role] --host [Host for Server-ID header]`

## Building

**Local**: `go mod download && go build`

**Pre-built, from Github**: https://github.com/bermannoah/simple-iam-vault-cli/releases

## Special thanks...

...to stormbeta for doing the work [here](https://github.com/stormbeta/snippets/tree/master/golang/vault-iam-auth).
