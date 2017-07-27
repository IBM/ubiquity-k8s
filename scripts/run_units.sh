#!/usr/bin/env bash
echo "Setting up ginkgo and gomega"
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega

echo "Starting unit tests...."
ginkgo -r -p --skip vendor -ldflags -s
