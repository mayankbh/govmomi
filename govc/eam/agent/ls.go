package agent

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/vmware/govmomi/eam"
	eamobject "github.com/vmware/govmomi/eam/object"
	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

func init() {
	cli.Register("eam.agent.ls", &ls{})
}

type ls struct {
	*flags.EAMFlags
}

// Process implements cli.Command.
func (cmd *ls) Process(ctx context.Context) error {
	cmd.ClientFlag.Process(ctx)
	cmd.OutputFlag.Process(ctx)
	return nil

}

// Register implements cli.Command.
func (cmd *ls) Register(ctx context.Context, f *flag.FlagSet) {
	// XXX Dedup with agency/Regster()?
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
		fmt.Printf("failed to get client")
		return err
	}
	resp, err := getAgents(ctx, vimClient)
	if err != nil {
		fmt.Printf("failed to get agents")
		return err
	}
	agents := agents(resp)
	cmd.OutputFlag.WriteResult(agents)
	return nil

}

func getAgents(ctx context.Context, client *vim25.Client) ([]*agentAndVM, error) {
	eamClient := eam.NewClient(client)
	eamHandle := eamobject.NewEsxAgentManager(eamClient, eam.EsxAgentManager)
	// XXX Dedup with agency/ls.go. Move to shared lib?
	agencies, err := eamHandle.Agencies(ctx)
	if err != nil {
		return nil, err
	}
	ret := []*agentAndVM{}
	for _, agency := range agencies {
		agents, err := agency.Agents(ctx)
		if err != nil {
			return nil, err
		}
		for _, agent := range agents {
			agentDetails, err := getAgent(ctx, client, agent)
			if err != nil {
				return nil, err
			}
			agentDetails.agency = agency
			ret = append(ret, agentDetails)
		}
	}
	return ret, nil
}

type agents []*agentAndVM

// Write implements flags.OutputWriter.
func (a agents) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)
	// XXX Dedup.
	printRow := func(s []string) {
		fmt.Fprintf(tw, s[0])
		for _, h := range s[1:] {
			fmt.Fprintf(tw, "\t%s", h)
		}
		fmt.Fprint(tw, "\n")
	}
	headers := []string{"Agency", "Agent ID", "Agent VM moref", "Agent VM name"}
	printRow(headers)
	for _, item := range a {
		value := []string{item.agency.Reference().Value, item.agent.Reference().Value, item.vm.Reference().Value, item.vm.Name}
		printRow(value)
	}
	return tw.Flush()
}

type agentAndVM struct {
	agent  eamobject.Agent
	agency eamobject.Agency
	vm     *mo.VirtualMachine
}

func getAgent(ctx context.Context, client *vim25.Client, agent eamobject.Agent) (*agentAndVM, error) {
	runtimeInfo, err := agent.Runtime(ctx)
	if err != nil {
		fmt.Printf("failed to get agent runtime")
		return nil, err
	}
	if runtimeInfo.Vm == nil {
		fmt.Printf("failed to get agent vm")
		return nil, nil
	}
	pc := property.DefaultCollector(client)
	vm := object.NewVirtualMachine(client, *runtimeInfo.Vm)
	v := &mo.VirtualMachine{}
	if err := pc.RetrieveOne(ctx, vm.Reference(), nil, v); err != nil {
		return nil, err
	}
	return &agentAndVM{vm: v, agent: agent}, nil
}
