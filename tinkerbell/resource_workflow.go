package tinkerbell

import (
	"context"
	"errors"
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateNotEmpty,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
			},
			"template": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateNotEmpty,
			},
		},
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.workflowClient

	req := workflow.CreateRequest{
		Template: d.Get("template").(string),
		Hardware: d.Get("hardwares").(string),
	}

	return diagsFromErr(retryOnTransientError(func() error {
		res, err := c.CreateWorkflow(ctx, &req)
		if err != nil {
			return fmt.Errorf("creating workflow: %w", err)
		}

		d.SetId(res.Id)

		return nil
	}))
}

func getWorkflow(ctx context.Context, c workflow.WorkflowServiceClient, uuid string) (*workflow.Workflow, error) {
	list, err := c.ListWorkflows(ctx, &workflow.Empty{})
	if err != nil {
		return nil, fmt.Errorf("getting all workflow entries: %w", err)
	}

	for {
		wf, err := list.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("receiving workflow entry: %w", err)
		}

		if wf == nil {
			return nil, fmt.Errorf("received empty workflow entry: %w", err)
		}

		if wf.GetId() == uuid {
			return wf, nil
		}
	}

	return nil, nil
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.workflowClient

	wf, err := getWorkflow(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("getting workflow %q: %w", d.Id(), err))
	}

	if wf == nil {
		d.SetId("")

		return nil
	}

	return nil
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.workflowClient

	wf, err := getWorkflow(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("getting workflow %q: %w", d.Id(), err))
	}

	if wf == nil {
		d.SetId("")

		return nil
	}

	req := workflow.GetRequest{
		Id: d.Id(),
	}

	if err := retryOnTransientError(func() error {
		_, err := c.DeleteWorkflow(ctx, &req)

		return err //nolint:wrapcheck
	}); err != nil {
		return diagsFromErr(fmt.Errorf("removing workflow %q: %w", d.Id(), err))
	}

	return nil
}
