default: install

install:
	go install .

build:
	go build -o terraform-provider-pihole .

# Requires: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
generate:
	tfplugindocs generate

.PHONY: default install build generate
