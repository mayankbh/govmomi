package agency

import (
	"context"
	"flag"

	"github.com/vmware/govmomi/eam"
	eammethods "github.com/vmware/govmomi/eam/methods"
	eamobject "github.com/vmware/govmomi/eam/object"
	eamtypes "github.com/vmware/govmomi/eam/types"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

func init() {
	cli.Register("eam.agency.destroy", &destroy{})
}

type destroy struct {
	*flags.EAMFlags
}

// Process implements cli.Command.
func (cmd *destroy) Process(ctx context.Context) error {
	return cmd.ClientFlag.Process(ctx)
}

var (
	agency string
)

// Register implements cli.Command.
func (cmd *destroy) Register(ctx context.Context, f *flag.FlagSet) {
	clientFlag, _ := flags.NewClientFlag(ctx)
	outputFlag, _ := flags.NewOutputFlag(ctx)
	cmd.EAMFlags = &flags.EAMFlags{}
	cmd.OutputFlag = outputFlag
	cmd.ClientFlag = clientFlag
	cmd.ClientFlag.Register(ctx, f)
	// XXX Positional arg?
	f.StringVar(&agency, "agency", "", "The EAM Agency to delete")
}

// Run implements cli.Command.
func (cmd *destroy) Run(ctx context.Context, f *flag.FlagSet) error {
	client, err := cmd.ClientFlag.Client()
	if err != nil {
		return err
	}

	eamClient := eam.NewClient(client)

	eamAgency := eamobject.NewAgency(eamClient, types.ManagedObjectReference{
		// XXX Const?
		Type:  "Agency",
		Value: agency,
	})
	req := &eamtypes.DestroyAgency{This: eamAgency.Reference()}

	// XXX Handle error?
	// No response.
	_, err = eammethods.DestroyAgency(ctx, eamClient, req)
	if err != nil {
		if soap.IsSoapFault(err) {
			// XXX Log soap fault?
		}
		return err
	}
	return nil
}
