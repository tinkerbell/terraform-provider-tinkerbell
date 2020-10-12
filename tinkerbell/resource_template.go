package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tinkerbell/tink/pkg"
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateNotEmpty,
			},
			"content": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateTemplate,
			},
		},
	}
}

func validateNotEmpty(m interface{}, p cty.Path) diag.Diagnostics {
	if m.(string) == "" {
		return diagsFromErr(fmt.Errorf("value must not be empty"))
	}

	return nil
}

func validateTemplate(m interface{}, p cty.Path) diag.Diagnostics {
	if m.(string) == "" {
		return diagsFromErr(fmt.Errorf("template content must not be empty"))
	}

	wf, err := pkg.ParseYAML([]byte(m.(string)))
	if err != nil {
		return diagsFromErr(fmt.Errorf("parsing template: %w", err))
	}

	if err := pkg.ValidateTemplate(wf); err != nil {
		return diagsFromErr(fmt.Errorf("validating template: %w", err))
	}

	return nil
}

func getTemplate(ctx context.Context, c template.TemplateClient, id string) (*template.WorkflowTemplate, error) {
	list, err := c.ListTemplates(ctx, &template.Empty{})
	if err != nil {
		return nil, fmt.Errorf("getting all template entries: %w", err)
	}

	for {
		t, err := list.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("receiving template entry: %w", err)
		}

		if t == nil {
			return nil, fmt.Errorf("received empty template entry: %w", err)
		}

		if t.GetId() == id {
			return t, nil
		}
	}

	return nil, nil
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.templateClient

	req := template.WorkflowTemplate{
		Name: d.Get("name").(string),
		Data: d.Get("content").(string),
	}

	return diagsFromErr(retryOnSerializationError(func() error {
		res, err := c.CreateTemplate(ctx, &req)
		if err != nil {
			return fmt.Errorf("creating template: %w", err)
		}

		d.SetId(res.Id)

		return nil
	}))
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.templateClient

	t, err := getTemplate(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if template exists: %w", err))
	}

	if t == nil {
		d.SetId("")

		return nil
	}

	req := template.GetRequest{
		Id: d.Id(),
	}

	t, err = c.GetTemplate(ctx, &req)
	if err != nil {
		return diagsFromErr(fmt.Errorf("getting template %q: %w", req.Id, err))
	}

	if err := d.Set("content", t.Data); err != nil {
		return diagsFromErr(fmt.Errorf("setting %q field: %w", "content", err))
	}

	return nil
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.templateClient

	t, err := getTemplate(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if template exists: %w", err))
	}

	if t == nil {
		d.SetId("")

		return nil
	}

	req := template.GetRequest{
		Id: d.Id(),
	}

	if err := retryOnSerializationError(func() error {
		_, err := c.DeleteTemplate(ctx, &req)

		return err
	}); err != nil {
		return diagsFromErr(fmt.Errorf("removing template: %w", err))
	}

	return nil
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.templateClient

	t, err := getTemplate(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if template exists: %w", err))
	}

	if t == nil {
		return diagsFromErr(fmt.Errorf("template %q do not exist: %w", d.Id(), err))
	}

	req := template.WorkflowTemplate{
		Id:   d.Id(),
		Name: d.Get("name").(string),
		Data: d.Get("content").(string),
	}

	if err := retryOnSerializationError(func() error {
		_, err := c.UpdateTemplate(ctx, &req)

		return err
	}); err != nil {
		return diagsFromErr(fmt.Errorf("updating template: %w", err))
	}

	return nil
}
