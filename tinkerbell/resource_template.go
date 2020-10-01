package tinkerbell

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tinkerbell/tink/protos/template"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		DeleteContext: resourceTemplateDelete,
		UpdateContext: resourceTemplateUpdate,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"content": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).TemplateClient

	req := template.WorkflowTemplate{
		Name: d.Get("name").(string),
		Data: d.Get("content").(string),
	}

	res, err := c.CreateTemplate(ctx, &req)
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating template failed: %w", err))
	}

	d.SetId(res.Id)

	return nil
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).TemplateClient

	// TODO: we should only do Get and distinguish fetch error from not found error
	// instead of iterating over all objects, as this doesn't scale.
	list, err := c.ListTemplates(ctx, &template.Empty{})
	if err != nil {
		return diagsFromErr(fmt.Errorf("listing templates failed: %w", err))
	}

	var tmp *template.WorkflowTemplate

	id := d.Id()
	found := false

	for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
		if tmp.Id == id {
			found = true

			break
		}
	}

	if err != nil && err != io.EOF {
		return diagsFromErr(fmt.Errorf("listing templates failed: %w", err))
	}

	if !found {
		d.SetId("")

		return nil
	}

	req := template.GetRequest{
		Id: d.Id(),
	}

	t, err := c.GetTemplate(ctx, &req)
	if err != nil {
		return diagsFromErr(fmt.Errorf("getting template failed: %w", err))
	}

	if err := d.Set("content", t.Data); err != nil {
		return diagsFromErr(fmt.Errorf("failed setting %q field: %w", "content", err))
	}

	return nil
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).TemplateClient

	req := template.GetRequest{
		Id: d.Id(),
	}

	if _, err := c.DeleteTemplate(ctx, &req); err != nil {
		return diagsFromErr(fmt.Errorf("removing template failed: %w", err))
	}

	return nil
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*tinkClient).TemplateClient

	req := template.WorkflowTemplate{
		Id:   d.Id(),
		Name: d.Get("name").(string),
		Data: d.Get("content").(string),
	}

	if _, err := c.UpdateTemplate(ctx, &req); err != nil {
		return diagsFromErr(fmt.Errorf("updating template failed: %w", err))
	}

	return nil
}
