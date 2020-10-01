package tinkerbell

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tinkerbell/tink/pkg"
	"github.com/tinkerbell/tink/protos/hardware"
)

func resourceHardware() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHardwareCreate,
		ReadContext:   resourceHardwareRead,
		DeleteContext: resourceHardwareDelete,
		UpdateContext: resourceHardwareCreate,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHardwareCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).HardwareClient

	hw := pkg.HardwareWrapper{}
	if err := json.Unmarshal([]byte(d.Get("data").(string)), &hw); err != nil {
		return diagsFromErr(fmt.Errorf("failed decoding 'data' as JSON: %w", err))
	}

	if hw.Hardware.Id == "" {
		return diagsFromErr(fmt.Errorf("ID is required in JSON data"))
	}

	if _, err := c.Push(ctx, &hardware.PushRequest{Data: hw.Hardware}); err != nil {
		return diagsFromErr(fmt.Errorf("failed pushing hardware data: %w", err))
	}

	d.SetId(hw.Hardware.Id)

	return nil
}

func resourceHardwareRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).HardwareClient

	// TODO: if error is not found, unset the ID to mark resource as non existent
	// instead of returning the error.
	hw, err := c.ByID(ctx, &hardware.GetRequest{Id: d.Id()})
	if err != nil {
		return diagsFromErr(fmt.Errorf("hardware with ID %q not found: %w", d.Id(), err))
	}

	b, err := json.Marshal(pkg.HardwareWrapper{Hardware: hw})
	if err != nil {
		return diagsFromErr(fmt.Errorf("serializing received hardware entry failed: %w", err))
	}

	if err := d.Set("data", string(b)); err != nil {
		return diagsFromErr(fmt.Errorf("failed setting %q field: %w", "data", err))
	}

	return nil
}

func resourceHardwareDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).HardwareClient

	req := hardware.DeleteRequest{
		Id: d.Id(),
	}

	if _, err := c.Delete(ctx, &req); err != nil {
		return diagsFromErr(fmt.Errorf("removing hardware failed: %w", err))
	}

	return nil
}
