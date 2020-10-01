package tinkerbell

import (
	"fmt"

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
		ResourcesMap: map[string]*schema.Resource{
			"tinkerbell_template": resourceTemplate(),
			"tinkerbell_workflow": resourceWorkflow(),
			"tinkerbell_hardware": resourceHardware(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	conn, err := client.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create tink client: %w", err)
	}

	return &tinkClient{
		TemplateClient: template.NewTemplateClient(conn),
		WorkflowClient: workflow.NewWorkflowSvcClient(conn),
		HardwareClient: hardware.NewHardwareServiceClient(conn),
	}, nil
}
