package tinkerbell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tinkerbell/tink/pkg"
	"github.com/tinkerbell/tink/protos/hardware"
)

const (
	dataAttribute = "data"
)

func resourceHardware() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHardwareCreate,
		ReadContext:   resourceHardwareRead,
		DeleteContext: resourceHardwareDelete,
		UpdateContext: resourceHardwareUpdate,
		Schema: map[string]*schema.Schema{
			dataAttribute: {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				ValidateDiagFunc: validateHardwareData,
			},
		},
	}
}

func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	return jsonBytesEqual(ob.Bytes(), nb.Bytes())
}

func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func validateHardwareData(m interface{}, p cty.Path) diag.Diagnostics {
	hw := pkg.HardwareWrapper{}

	if err := json.Unmarshal([]byte(m.(string)), &hw); err != nil {
		return diagsFromErr(fmt.Errorf("failed decoding 'data' as JSON: %w", err))
	}

	if hw.Hardware.Id == "" {
		return diagsFromErr(fmt.Errorf("ID is required in JSON data"))
	}

	return nil
}

func resourceHardwareCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.hardwareClient

	hw := pkg.HardwareWrapper{}

	// We can skip error checking here, validate function should already validate it.
	_ = json.Unmarshal([]byte(d.Get(dataAttribute).(string)), &hw)

	h, err := getHardware(ctx, c, hw.Hardware.Id)
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if hardware ID %q already exists: %w", hw.Hardware.Id, err))
	}

	if h != nil {
		return diagsFromErr(fmt.Errorf("hardware ID %q already exists", hw.Hardware.Id))
	}

	if _, err := c.Push(ctx, &hardware.PushRequest{Data: hw.Hardware}); err != nil {
		return diagsFromErr(fmt.Errorf("pushing hardware data: %w", err))
	}

	d.SetId(hw.Hardware.Id)

	return nil
}

func resourceHardwareUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.hardwareClient

	hw := pkg.HardwareWrapper{}

	// We can skip error checking here, validate function should already validate it.
	_ = json.Unmarshal([]byte(d.Get(dataAttribute).(string)), &hw)

	h, err := getHardware(ctx, c, hw.Hardware.Id)
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if hardware ID %q already exists: %w", hw.Hardware.Id, err))
	}

	if h == nil {
		return diagsFromErr(fmt.Errorf("hardware ID %q does not exist", hw.Hardware.Id))
	}

	if _, err := c.Push(ctx, &hardware.PushRequest{Data: hw.Hardware}); err != nil {
		return diagsFromErr(fmt.Errorf("pushing hardware data: %w", err))
	}

	d.SetId(hw.Hardware.Id)

	return nil
}

func getHardware(ctx context.Context, c hardware.HardwareServiceClient, uuid string) (*hardware.Hardware, error) {
	list, err := c.All(ctx, &hardware.Empty{})
	if err != nil {
		return nil, fmt.Errorf("getting all hardware entries: %w", err)
	}

	for {
		hw, err := list.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("receiving hardware entry: %w", err)
		}

		if hw == nil {
			return nil, fmt.Errorf("received empty hardware entry: %w", err)
		}

		if hw.GetId() == uuid {
			return hw, nil
		}
	}

	return nil, nil
}

func resourceHardwareRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.hardwareClient

	h, err := getHardware(ctx, c, d.Id())
	if err != nil {
		return diagsFromErr(fmt.Errorf("checking if hardware %q exists: %w", d.Id(), err))
	}

	if h == nil {
		d.SetId("")

		return nil
	}

	b, err := json.Marshal(pkg.HardwareWrapper{Hardware: h})
	if err != nil {
		return diagsFromErr(fmt.Errorf("serializing received hardware entry failed: %w", err))
	}

	if err := d.Set(dataAttribute, string(b)); err != nil {
		return diagsFromErr(fmt.Errorf("failed setting %q field: %w", dataAttribute, err))
	}

	return nil
}

func resourceHardwareDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc, err := m.(*tinkClientConfig).New()
	if err != nil {
		return diagsFromErr(fmt.Errorf("creating Tink client: %w", err))
	}

	c := tc.hardwareClient

	req := hardware.DeleteRequest{
		Id: d.Id(),
	}

	if _, err := c.Delete(ctx, &req); err != nil {
		return diagsFromErr(fmt.Errorf("removing hardware failed: %w", err))
	}

	return nil
}
