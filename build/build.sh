#!/usr/bin/env bash

# There's no way to build musl binaries with github
# so we need to compile it inside a container
# and manually push the binary to github

VERSION=${VERSION:=0.18}

DOCKERFILE=${DOCKERFILE:-Dockerfile}

docker build -f ${DOCKERFILE} -t devfh/simple-iam-vault-cli:${VERSION} ..
docker create -it --name simple-iam-vault-cli-build devfh/simple-iam-vault-cli:${VERSION} bash
docker cp simple-iam-vault-cli-build:/go/bin/simple-iam-vault-cli .
docker rm -f simple-iam-vault-cli-build