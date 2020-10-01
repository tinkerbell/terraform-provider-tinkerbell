package tinkerbell

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tinkerbell/tink/protos/workflow"
)

func resourceWorkflow() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkflowCreate,
		ReadContext:   resourceWorkflowRead,
		DeleteContext: resourceWorkflowDelete,
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

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).WorkflowClient

	req := workflow.CreateRequest{
		Template: d.Get("template").(string),
		Hardware: d.Get("hardwares").(string),
	}

	res, err := c.CreateWorkflow(ctx, &req)
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating workflow failed: %w", err))
	}

	d.SetId(res.Id)

	return nil
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).WorkflowClient

	// TODO: we should only do Get and distinguish fetch error from not found error
	// instead of iterating over all objects, as this doesn't scale.
	list, err := c.ListWorkflows(ctx, &workflow.Empty{})
	if err != nil {
		return diagsFromErr(fmt.Errorf("listing workflows failed: %w", err))
	}

	var tmp *workflow.Workflow

	id := d.Id()
	found := false

	for tmp, err = list.Recv(); err == nil && tmp.Id != ""; tmp, err = list.Recv() {
		if tmp.Id == id {
			found = true

			break
		}
	}

	if err != nil && err != io.EOF {
		return diagsFromErr(fmt.Errorf("listing workflows failed: %w", err))
	}

	if !found {
		d.SetId("")

		return nil
	}

	return nil
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).WorkflowClient

	req := workflow.GetRequest{
		Id: d.Id(),
	}

	if _, err := c.DeleteWorkflow(ctx, &req); err != nil {
		return diagsFromErr(fmt.Errorf("removing workflow failed: %w", err))
	}

	return nil
}
