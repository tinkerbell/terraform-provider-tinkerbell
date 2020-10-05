package tinkerbell

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

type tinkClient struct {
	TemplateClient template.TemplateClient
	WorkflowClient workflow.WorkflowSvcClient
	HardwareClient hardware.HardwareServiceClient
}

// Provider returns the Tinkerbell terraform provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"grpc_authority": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Equivalent of TINKERBELL_GRPC_AUTHORITY environment variable.",
			},
			"cert_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Equivalent of TINKERBELL_CERT_URL environment variable.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tinkerbell_template": resourceTemplate(),
			"tinkerbell_workflow": resourceWorkflow(),
			"tinkerbell_hardware": resourceHardware(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	if v := d.Get("grpc_authority").(string); v != "" {
		if err := os.Setenv("TINKERBELL_GRPC_AUTHORITY", v); err != nil {
			return nil, fmt.Errorf("setting TINKERBELL_GRPC_AUTHORITY environment variable: %w", err)
		}
	}

	if v := d.Get("cert_url").(string); v != "" {
		if err := os.Setenv("TINKERBELL_CERT_URL", v); err != nil {
			return nil, fmt.Errorf("setting TINKERBELL_CERT_URL environment variable: %w", err)
		}
	}

	conn, err := client.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("creating tink client: %w", err)
	}

	return &tinkClient{
		TemplateClient: template.NewTemplateClient(conn),
		WorkflowClient: workflow.NewWorkflowSvcClient(conn),
		HardwareClient: hardware.NewHardwareServiceClient(conn),
	}, nil
}
