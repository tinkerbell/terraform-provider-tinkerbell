# Tinkerbell Provider

The Tinkerbell provider allows to create [Tinkerbell](https://tinkerbell.org/) hardware entried, templates and workflows in a declarative way.

## Example Usage

```hcl
terraform {
  required_providers {
    tinkerbell = {
      source  = "tinkerbell/tinkerbell"
      version = "0.1.0"
    }
  }
}

provider "tinkerbell" {
  grpc_authority = "127.0.0.1:42113"
  cert_url       = "http://127.0.0.1:42114/cert"
}

resource "tinkerbell_hardware" "foo" {
  data = <<EOF
{
  "id": "2bd4b2b3-3104-4f67-8b5c-3d208d9cd1cd",
  "metadata": {
    "facility": {
      "facility_code": "ewr1",
      "plan_slug": "c2.medium.x86",
      "plan_version_slug": ""
    },
    "instance": {},
    "state": "provisioning"
  },
  "network": {
    "interfaces": [
      {
        "dhcp": {
          "arch": "x86_64",
          "ip": {
            "address": "192.168.1.5",
            "gateway": "192.168.1.1",
            "netmask": "255.255.255.248"
          },
          "mac": "ff:ff:ff:ff:ff:ff"
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
EOF
}

resource "tinkerbell_template" "foo" {
  name    = "foo"
  content = <<EOF
version: "0.1"
name: ubuntu_provisioning
global_timeout: 6000
tasks:
  - name: "os-installation"
    worker: "{{.device_1}}"
    volumes:
      - /dev:/dev
      - /dev/console:/dev/console
      - /lib/firmware:/lib/firmware:ro
    environment:
      MIRROR_HOST: <MIRROR_HOST_IP>
    actions:
      - name: "disk-wipe"
        image: disk-wipe
        timeout: 90
      - name: "disk-partition"
        image: disk-partition
        timeout: 600
        environment:
          MIRROR_HOST: <MIRROR_HOST_IP>
        volumes:
          - /statedir:/statedir
      - name: "install-root-fs"
        image: install-root-fs
        timeout: 600
      - name: "install-grub"
        image: install-grub
        timeout: 600
        volumes:
          - /statedir:/statedir
EOF
}

resource "tinkerbell_workflow" "foo" {
	template  = tinkerbell_template.foo.id
  hardwares = hardwares = <<EOF
{"device_1":"ff:ff:ff:ff:ff:ff"}
EOF
}
```

## Argument Reference

* `grpc_authority` - (Optional) Equivalent of TINKERBELL_GRPC_AUTHORITY environment variable.

* `cert_url` - (Optional) Equivalent of TINKERBELL_CERT_URL environment variable.
