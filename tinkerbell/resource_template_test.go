package tinkerbell

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccTemplate(name, content string) string {
	return fmt.Sprintf(`
resource "tinkerbell_template" "foo" {
	name    = "%s"
	content = <<EOF
%s
EOF
}
`, name, content)
}

func testAccTemplateContent(timeout int) string {
	return fmt.Sprintf(`
version: "0.1"
name: ubuntu_provisioning
global_timeout: %d
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
`, timeout)
}

func TestAccTemplate_create(t *testing.T) {
	name := newUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplate(name, testAccTemplateContent(1)),
			},
		},
	})
}

func TestAccTemplate_detectChanges(t *testing.T) {
	name := newUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplate(name, testAccTemplateContent(1)),
			},
			{
				Config:             testAccTemplate(name, testAccTemplateContent(2)),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccTemplate_update(t *testing.T) {
	name := newUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplate(name, testAccTemplateContent(1)),
			},
			{
				Config: testAccTemplate(name, testAccTemplateContent(2)),
			},
		},
	})
}

func TestAccTemplate_validateData(t *testing.T) {
	name := newUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccTemplate(name, `"foo`),
				ExpectError: regexp.MustCompile(`parsing template`),
			},
			{
				Config:      testAccTemplate(name, `foo: bar`),
				ExpectError: regexp.MustCompile(`parsing template`),
			},
		},
	})
}
