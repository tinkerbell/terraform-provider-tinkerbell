provider "tinkerbell" {}

resource "tinkerbell_template" "test" {
  name    = "bar"
  content = <<EOF
version: '0.1'
name: flatcar-install
global_timeout: 1800
tasks:
- name: "flatcar-install"
  worker: "{{index .Targets "machine1" "mac_addr"}}"
  volumes:
  - /dev:/dev
  - /statedir:/statedir
  actions:
  - name: "dump-ignition"
    image: alpine
    command:
    - sh
    - -c
    - echo '{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.2.0"},"networkd":{},"passwd":{"users":[{"name":"core","sshAuthorizedKeys":[]}]},"storage":{},"systemd":{}}' > /statedir/ignition.json
  - name: "flatcar-install"
    image: flatcar-install
    command:
    - -d
    - /dev/sda
    - -i
    - /statedir/ignition.json
    - -b
    - http://192.168.50.3/misc/flatcar/stable/amd64-usr
  - name: "reboot"
    image: alpine
    command:
    - sh
    - -c
    - 'echo 1 > /proc/sys/kernel/sysrq; echo b > /proc/sysrq-trigger'
EOF
}

resource "tinkerbell_target" "test" {
  data = <<EOF
{"targets": {"machine1": {"mac_addr": "08:00:27:ea:b7:89"}}}
EOF
}

resource "tinkerbell_workflow" "test" {
  target   = tinkerbell_target.test.id
  template = tinkerbell_template.test.id
}
