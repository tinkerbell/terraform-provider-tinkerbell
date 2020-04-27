package tinkerbell

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/tinkerbell/tink/protos/template"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Delete: resourceTemplateDelete,
		Update: resourceTemplateUpdate,
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

func resourceTemplateCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TemplateClient

	req := template.WorkflowTemplate{
		Name: d.Get("name").(string),
		Data: []byte(d.Get("content").(string)),
	}

	res, err := c.CreateTemplate(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("creating template failed: %w", err)
	}

	d.SetId(res.Id)

	return nil
}

func resourceTemplateRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TemplateClient

	// TODO: we should only do Get and distinguish fetch error from not found error
	// instead of iterating over all objects, as this doesn't scale.
	list, err := c.ListTemplates(context.Background(), &template.Empty{})
	if err != nil {
		return fmt.Errorf("listing templates failed: %w", err)
	}

	var tmp *template.WorkflowTemplate
	err = nil

	id := d.Id()
	found := false

	for tmp, err = list.Recv(); err == nil && tmp.Name != ""; tmp, err = list.Recv() {
		if tmp.Id == id {
			found = true

			break
		}
	}

	if err != nil && err != io.EOF {
		return fmt.Errorf("listing templates failed: %w", err)
	}

	if !found {
		d.SetId("")

		return nil
	}

	req := template.GetRequest{
		Id: d.Id(),
	}

	t, err := c.GetTemplate(context.Background(), &req)
	if err != nil {
		return fmt.Errorf("getting template failed: %w", err)
	}

	if err := d.Set("content", string(t.Data)); err != nil {
		return fmt.Errorf("failed setting %q field: %w", "content", err)
	}

	return nil
}

func resourceTemplateDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TemplateClient

	req := template.GetRequest{
		Id: d.Id(),
	}

	if _, err := c.DeleteTemplate(context.Background(), &req); err != nil {
		return fmt.Errorf("removing template failed: %w", err)
	}

	return nil
}

func resourceTemplateUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*TinkClient).TemplateClient

	req := template.WorkflowTemplate{
		Id:   d.Id(),
		Name: d.Get("name").(string),
		Data: []byte(d.Get("content").(string)),
	}

	if _, err := c.UpdateTemplate(context.Background(), &req); err != nil {
		return fmt.Errorf("updating template failed: %w", err)
	}

	return nil
}
