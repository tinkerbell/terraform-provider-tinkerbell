package tinkerbell

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/tinkerbell/tink/protos/target"
)

func resourceTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceTargetCreate,
		Read:   resourceTargetRead,
		Delete: resourceTargetDelete,
		Update: resourceTargetUpdate,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTargetCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TargetClient

	req := target.PushRequest{
		Data: strings.TrimSpace(d.Get("data").(string)),
	}

	id, err := c.CreateTargets(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("creating target failed: %w", err)
	}

	d.SetId(id.Uuid)

	return nil
}

func resourceTargetRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TargetClient

	// TODO: we should only do Get and distinguish fetch error from not found error
	// instead of iterating over all objects, as this doesn't scale.
	list, err := c.ListTargets(context.Background(), &target.Empty{})
	if err != nil {
		return fmt.Errorf("listing targets failed: %w", err)
	}

	var tmp *target.TargetList
	err = nil

	id := d.Id()
	found := false

	for tmp, err = list.Recv(); err == nil && tmp.Data != ""; tmp, err = list.Recv() {
		if tmp.ID == id {
			found = true

			break
		}
	}

	if err != nil && err != io.EOF {
		return fmt.Errorf("listing targets failed: %w", err)
	}

	if !found {
		d.SetId("")

		return nil
	}

	req := target.GetRequest{
		ID: d.Id(),
	}

	t, err := c.TargetByID(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("getting target failed: %w", err)
	}

	// TODO: this currently produces constant diff because of whitespace for some reason.
	if err := d.Set("data", string(t.JSON)); err != nil {
		return fmt.Errorf("failed setting %q field: %w", "content", err)
	}

	return nil
}

func resourceTargetDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TargetClient

	req := target.GetRequest{
		ID: d.Id(),
	}

	if _, err := c.DeleteTargetByID(context.Background(), &req); err != nil {
		return fmt.Errorf("removing target failed: %w", err)
	}

	return nil
}

func resourceTargetUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TargetClient

	req := target.UpdateRequest{
		ID:   d.Id(),
		Data: d.Get("data").(string),
	}

	if _, err := c.UpdateTargetByID(context.Background(), &req); err != nil {
		return fmt.Errorf("updating target failed: %w", err)
	}

	return nil
}
