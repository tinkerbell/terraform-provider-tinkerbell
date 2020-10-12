package tinkerbell

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccWorkflow(t *testing.T, id int) string {
	name := newUUID(t)
	rMAC := newMAC(t)
	resourceName := fmt.Sprintf("foo%d", id)

	return fmt.Sprintf(`
%s

%s

resource "tinkerbell_workflow" "%s" {
	template  = tinkerbell_template.a%s.id
	hardwares = <<EOF
{"device_1":"%s"}
EOF

	depends_on = [
		tinkerbell_hardware.%s,
	]
}
`,
		testAccHardware(testAccHardwareConfig(name, rMAC), resourceName),
		testAccTemplate(name, testAccTemplateContent(1)),
		resourceName,
		name,
		rMAC,
		resourceName,
	)
}

func TestAccWorkflow_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflow(t, 0),
			},
		},
	})
}

func TestAccWorkflow_parallel(t *testing.T) {
	config := ""
	for i := 0; i < 10; i++ {
		config += testAccWorkflow(t, i)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
		},
	})
}
