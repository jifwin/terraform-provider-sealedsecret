#!/bin/bash

set -e

FILE="terraform-provider-sealedsecret"
cd ..
go build -o "${FILE}"
mv "${FILE}" ~/.terraform.d/plugins/terraform.example.com/local/sealedsecret/0.0.1/linux_amd64
cd examples/

rm -rf .terraform || true
rm .terraform.lock.hcl || true
rm terraform.tfstate || true

terraform init
terraform apply
