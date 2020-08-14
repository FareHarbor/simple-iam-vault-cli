FROM alpine:latest

RUN apk add --no-cache git make musl-dev go

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

COPY . ${GOPATH}/src

# Install Glide
RUN go get -u github.com/Masterminds/glide/...

WORKDIR $GOPATH/src

RUN go mod download && go build && \
    mkdir -p $GOPATH/bin && mv simple-iam-vault-cli $GOPATH/bin/simple-iam-vault-cli

CMD ["simple-iam-vault-cli"]
