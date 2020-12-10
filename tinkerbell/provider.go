package tinkerbell

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/template"
	"github.com/tinkerbell/tink/protos/workflow"
)

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

type tinkClientConfig struct {
	providerConfig *schema.ResourceData
	client         *tinkClient
	clientMutex    sync.Mutex
}

type tinkClient struct {
	templateClient template.TemplateServiceClient
	workflowClient workflow.WorkflowServiceClient
	hardwareClient hardware.HardwareServiceClient
}

func (tc *tinkClientConfig) New() (*tinkClient, error) {
	tc.clientMutex.Lock()
	defer tc.clientMutex.Unlock()

	if tc.client != nil {
		return tc.client, nil
	}

	if grpcAuthority := tc.providerConfig.Get("grpc_authority").(string); grpcAuthority != "" {
		if err := os.Setenv("TINKERBELL_GRPC_AUTHORITY", grpcAuthority); err != nil {
			return nil, fmt.Errorf("setting TINKERBELL_GRPC_AUTHORITY environment variable: %w", err)
		}
	}

	if certURL := tc.providerConfig.Get("cert_url").(string); certURL != "" {
		if err := os.Setenv("TINKERBELL_CERT_URL", certURL); err != nil {
			return nil, fmt.Errorf("setting TINKERBELL_CERT_URL environment variable: %w", err)
		}
	}

	conn, err := client.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("creating tink client: %w", err)
	}

	tc.client = &tinkClient{
		templateClient: template.NewTemplateServiceClient(conn),
		workflowClient: workflow.NewWorkflowServiceClient(conn),
		hardwareClient: hardware.NewHardwareServiceClient(conn),
	}

	return tc.client, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return &tinkClientConfig{
		providerConfig: d,
	}, nil
}
