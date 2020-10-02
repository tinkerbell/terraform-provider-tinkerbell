package tinkerbell

import (
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
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("Internal provider validation: %v", err)
	}
}
