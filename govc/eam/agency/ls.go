package agency

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/vmware/govmomi/eam"
	eamobject "github.com/vmware/govmomi/eam/object"
	eamtypes "github.com/vmware/govmomi/eam/types"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
)

func init() {
	cli.Register("eam.agency.ls", &ls{})
}

type ls struct {
	*flags.EAMFlags
}

type agencies []agencyAndConfig

type agencyAndConfig struct {
	agency eamobject.Agency
	config *eamtypes.AgencyConfigInfo
}

// List all agencies.
// Get each one's config.
// Return list.
func getAgencies(ctx context.Context, client *eam.Client) ([]agencyAndConfig, error) {
	eamInstance := eamobject.NewEsxAgentManager(client, eam.EsxAgentManager)
	resp, err := eamInstance.Agencies(ctx)
	if err != nil {
		return nil, err
	}
	ret := []agencyAndConfig{}
	for _, item := range resp {
		config, err := item.Config(ctx)
		if err != nil {
			return nil, err
		}
		ret = append(ret, agencyAndConfig{agency: item, config: config.GetAgencyConfigInfo()})
	}
	return ret, nil
}

// Write implements flags.OutputWriter.
func (a agencies) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)
	// Headers
	printRow := func(s []string) {
		fmt.Fprintf(tw, s[0])
		for _, h := range s[1:] {
			fmt.Fprintf(tw, "\t%s", h)
		}
		fmt.Fprint(tw, "\n")
	}
	agencyHeaders := []string{"Agency ID", "Agency Name", "Agent name"}
	printRow(agencyHeaders)

	for _, i := range a {
		fields := []string{i.agency.Reference().Value, i.config.GetAgencyConfigInfo().AgencyName, i.config.GetAgencyConfigInfo().AgentName}
		printRow(fields)
	}
	tw.Flush()
	return nil
}

// Process implements cli.Command.
func (cmd *ls) Process(ctx context.Context) error {
	cmd.ClientFlag.Process(ctx)
	cmd.OutputFlag.Process(ctx)
	return nil
}

// Register implements cli.Command.
func (cmd *ls) Register(ctx context.Context, f *flag.FlagSet) {
	clientFlag, _ := flags.NewClientFlag(ctx)
	outputFlag, _ := flags.NewOutputFlag(ctx)
	cmd.EAMFlags = &flags.EAMFlags{}
	cmd.ClientFlag = clientFlag
	cmd.OutputFlag = outputFlag
	cmd.ClientFlag.Register(ctx, f)
	cmd.OutputFlag.Register(ctx, f)
}

// Run implements cli.Command.
func (cmd *ls) Run(ctx context.Context, f *flag.FlagSet) error {
	vimClient, err := cmd.Client()
	if err != nil {
		return err
	}
	eamClient := eam.NewClient(vimClient)
	resp, err := getAgencies(ctx, eamClient)
	if err != nil {
		return err
	}
	agencies := agencies(resp)
	cmd.OutputFlag.WriteResult(agencies)
	return nil
}
