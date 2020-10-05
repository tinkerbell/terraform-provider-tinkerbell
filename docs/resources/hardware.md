# Hardware Resource

This resource allows to create Tinkerbell [hardware](https://docs.tinkerbell.org/about/workflows/) data.

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
```

## Argument Reference

* `data` - (Required) JSON formatted hardware data. See Tinkerbell [documentation](https://docs.tinkerbell.org/about/hardware-data/) for available fields and their documentation.
