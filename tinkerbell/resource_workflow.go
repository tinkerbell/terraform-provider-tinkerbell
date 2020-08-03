package tinkerbell

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/tinkerbell/tink/protos/workflow"
)

func resourceWorkflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkflowCreate,
		Read:   resourceWorkflowRead,
		Delete: resourceWorkflowDelete,
		Schema: map[string]*schema.Schema{
			"hardwares": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkflowCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).WorkflowClient

	req := workflow.CreateRequest{
		Template: d.Get("template").(string),
		Hardware: d.Get("hardwares").(string),
	}

	res, err := c.CreateWorkflow(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("creating workflow failed: %w", err)
	}

	d.SetId(res.Id)

	return nil
}

func resourceWorkflowRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).WorkflowClient

	// TODO: we should only do Get and distinguish fetch error from not found error
	// instead of iterating over all objects, as this doesn't scale.
	list, err := c.ListWorkflows(context.Background(), &workflow.Empty{})
	if err != nil {
		return fmt.Errorf("listing workflows failed: %w", err)
	}

	var tmp *workflow.Workflow
	err = nil

	id := d.Id()
	found := false

	for tmp, err = list.Recv(); err == nil && tmp.Id != ""; tmp, err = list.Recv() {
		if tmp.Id == id {
			found = true

			break
		}
	}

	if err != nil && err != io.EOF {
		return fmt.Errorf("listing workflows failed: %w", err)
	}

	if !found {
		d.SetId("")

		return nil
	}

	return nil
}

func resourceWorkflowDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).WorkflowClient

	req := workflow.GetRequest{
		Id: d.Id(),
	}

	if _, err := c.DeleteWorkflow(context.Background(), &req); err != nil {
		return fmt.Errorf("removing workflow failed: %w", err)
	}

	return nil
}
