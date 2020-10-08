package tinkerbell

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// From https://stackoverflow.com/a/21027407/2974814
func newMAC(t *testing.T) string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("Generating MAC address: %v", err)
	}
	// Set the local bit
	buf[0] |= 2

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

func testAccHardwareConfig(uuid string, mac string) string {
	return fmt.Sprintf(`
{
  "id": "%s",
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
          "mac": "%s"
        },
        "netboot": {
          "allow_pxe": true,
          "allow_workflow": true
        }
      }
    ]
  }
}
`, uuid, mac)
}

func testAccHardware(data string) string {
	return fmt.Sprintf(`
resource "tinkerbell_hardware" "foo" {
	data = <<EOF
%s
EOF
}
`, data)
}

func newUUID(t *testing.T) string {
	i, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("Generating UUID: %v", err)
	}

	return i.String()
}

func TestAccHardware_create(t *testing.T) {
	rUUID := newUUID(t)
	rMAC := newMAC(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, rMAC)),
			},
		},
	})
}

func TestAccHardware_detectChanges(t *testing.T) {
	rUUID := newUUID(t)
	rMAC := newMAC(t)
	nMAC := newMAC(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, rMAC)),
			},
			{
				Config:             testAccHardware(testAccHardwareConfig(rUUID, nMAC)),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccHardware_update(t *testing.T) {
	rUUID := newUUID(t)
	rMAC := newMAC(t)
	nMAC := newMAC(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, rMAC)),
			},
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, nMAC)),
			},
		},
	})
}

func TestAccHardware_updateUUID(t *testing.T) {
	rUUID := newUUID(t)
	nUUID := newUUID(t)
	rMAC := newMAC(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, rMAC)),
			},
			{
				Config: testAccHardware(testAccHardwareConfig(nUUID, rMAC)),
			},
		},
	})
}

func TestAccHardware_ignoreWhitespace(t *testing.T) {
	rUUID := newUUID(t)
	rMAC := newMAC(t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccHardware(testAccHardwareConfig(rUUID, rMAC)),
			},
			{
				Config:             testAccHardware(fmt.Sprintf("%s\n", testAccHardwareConfig(rUUID, rMAC))),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccHardware_validateData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccHardware("bad json"),
				ExpectError: regexp.MustCompile(`failed decoding 'data' as JSON`),
			},
		},
	})
}
