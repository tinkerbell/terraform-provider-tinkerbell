package tinkerbell

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/util"
)

func resourceHardware() *schema.Resource {
	return &schema.Resource{
		Create: resourceHardwareCreate,
		Read:   resourceHardwareRead,
		Delete: resourceHardwareDelete,
		Update: resourceHardwareCreate,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHardwareCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).HardwareClient

	hw := util.HardwareWrapper{}
	if err := json.Unmarshal([]byte(d.Get("data").(string)), &hw); err != nil {
		return fmt.Errorf("failed decoding 'data' as JSON: %w", err)
	}

	if hw.Hardware.Id == "" {
		return fmt.Errorf("ID is required in JSON data")
	}

	if _, err := c.Push(context.Background(), &hardware.PushRequest{Data: hw.Hardware}); err != nil {
		return fmt.Errorf("failed pushing hardware data: %w", err)
	}

	d.SetId(hw.Hardware.Id)

	return nil
}

func resourceHardwareRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).HardwareClient

	// TODO: if error is not found, unset the ID to mark resource as non existent
	// instead of returning the error.
	hw, err := c.ByID(context.Background(), &hardware.GetRequest{Id: d.Id()})
	if err != nil {
		return fmt.Errorf("hardware with ID %q not found: %w", d.Id(), err)
	}

	b, err := json.Marshal(util.HardwareWrapper{Hardware: hw})
	if err != nil {
		return fmt.Errorf("serializing received hardware entry failed: %w", err)
	}

	if err := d.Set("data", string(b)); err != nil {
		return fmt.Errorf("failed setting %q field: %w", "data", err)
	}

	return nil
}

func resourceHardwareDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).HardwareClient

	req := hardware.DeleteRequest{
		Id: d.Id(),
	}

	if _, err := c.Delete(context.Background(), &req); err != nil {
		return fmt.Errorf("removing hardware failed: %w", err)
	}

	return nil
}
