# Template Resource

This resource allows to create Tinkerbell [templates](https://docs.tinkerbell.org/about/templates/).

## Example Usage

```hcl
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
```

## Argument Reference

* `name` - (Required) Template name.
* `content` - (Requires) Template content in YAML format. See Tinkerbell [documentation](https://docs.tinkerbell.org/about/templates/) for more details.
