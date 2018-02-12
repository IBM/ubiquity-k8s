FROM golang:1.9.1
WORKDIR /go/src/github.com/IBM/ubiquity-k8s/
COPY . .
RUN go get -v github.com/Masterminds/glide
RUN glide up --strip-vendor
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity-k8s-flex cmd/flex/main/cli.go


FROM alpine:latest
RUN apk --no-cache add ca-certificates
ENV UBIQUITY_PLUGIN_VERIFY_CA=/var/lib/ubiquity/ssl/public/ubiquity-trusted-ca.crt
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/ubiquity-k8s-flex .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/deploy/k8s_deployments/setup_flex.sh .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/LICENSE .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/scripts/notices_file_for_ibm_storage_enabler_for_containers ./NOTICES
