package tinkerbell

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func diagsFromErr(err error) diag.Diagnostics {
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
