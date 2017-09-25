FROM golang:1.9.0
WORKDIR /go/src/github.com/IBM/ubiquity-k8s/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up --strip-vendor
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity-k8s-flex cmd/flex/main/cli.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/ubiquity-k8s-flex .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/deploy/k8s_deployments/setup_flex.sh .