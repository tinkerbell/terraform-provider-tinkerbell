package tinkerbell

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:gochecknoglobals
var testAccProviders map[string]*schema.Provider

//nolint:gochecknoglobals
var testAccProvider *schema.Provider

//nolint:gochecknoinits
func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"tinkerbell": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	t.Parallel()

	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("Internal provider validation: %v", err)
	}
}

// testAccPreCheck validates the necessary test environment variables exist.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("TINKERBELL_GRPC_AUTHORITY"); v == "" {
		t.Fatal("TINKERBELL_GRPC_AUTHORITY must be set for acceptance tests")
	}

	if v := os.Getenv("TINKERBELL_CERT_URL"); v == "" {
		t.Fatal("TINKERBELL_CERT_URL must be set for acceptance tests")
	}
}
