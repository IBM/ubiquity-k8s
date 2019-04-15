FROM golang:1.9.1
WORKDIR /go/src/github.com/IBM/ubiquity-k8s/
RUN go get -v github.com/Masterminds/glide
ADD glide.yaml .
RUN glide up --strip-vendor
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity-k8s-flex cmd/flex/main/cli.go


FROM alpine:3.8
RUN apk --no-cache add ca-certificates=20171114-r3
ENV UBIQUITY_PLUGIN_VERIFY_CA=/var/lib/ubiquity/ssl/public/ubiquity-trusted-ca.crt
WORKDIR /root/
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/ubiquity-k8s-flex .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/scripts/setup_flex.sh .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/scripts/health_check.sh .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/LICENSE .
COPY --from=0 /go/src/github.com/IBM/ubiquity-k8s/scripts/notices_file_for_ibm_storage_enabler_for_containers ./NOTICES
