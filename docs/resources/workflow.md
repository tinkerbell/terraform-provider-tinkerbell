# Workflow Resource

This resource allows to create Tinkerbell [workflows](https://docs.tinkerbell.org/about/workflows/).

## Example Usage

```hcl
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

* `template` - (Required) Template ID to use.
* `hardwares` - (Requires) JSON formatted map of hardwares to create a workflow for, where key is device name and value is MAC address of desired hardware. See Tinkerbell [documentation](https://docs.tinkerbell.org/about/workflows/) for more details.
