package tinkerbell

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccWorkflow(t *testing.T) string {
	name := newUUID(t)
	rMAC := newMAC(t)

	return fmt.Sprintf(`
%s

%s

resource "tinkerbell_workflow" "foo" {
	template  = tinkerbell_template.foo.id
	hardwares = <<EOF
{"device_1":"%s"}
EOF

	depends_on = [
		tinkerbell_hardware.foo,
	]
}
`,
		testAccHardware(testAccHardwareConfig(name, rMAC)),
		testAccTemplate(name, testAccTemplateContent(1)),
		rMAC,
	)
}

func TestAccWorkflow_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflow(t),
			},
		},
	})
}
