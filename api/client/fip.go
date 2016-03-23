package client

import (
	"fmt"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// CmdFip is the parent subcommand for all fip commands
//
// Usage: docker fip <COMMAND> [OPTIONS]
func (cli *DockerCli) CmdFip(args ...string) error {
	cmd := Cli.Subcmd("fip", []string{"COMMAND [OPTIONS]"}, fipUsage(), false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	cmd.Usage()
	return err
}

// CmdNetworkCreate creates a new fip with a given name
//
// Usage: docker fip create [OPTIONS] COUNT
func (cli *DockerCli) CmdFipAllocate(args ...string) error {
	cmd := Cli.Subcmd("fip allocate", []string{"COUNT"}, "Creates some new floating IPs by the user", false)

	cmd.Require(flag.Exact, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	fips, err := cli.client.FipAllocate(cmd.Arg(0))
	if err != nil {
		return err
	}
	for _, ip := range fips {
		fmt.Fprintf(cli.out, "%s\n", ip)
	}
	return nil
}

// CmdFipRelease deletes one or more fips
//
// Usage: docker fip release FIP [FIP...]
func (cli *DockerCli) CmdFipRelease(args ...string) error {
	cmd := Cli.Subcmd("fip release", []string{"FIP [FIP...]"}, "Release one or more fips", false)
	cmd.Require(flag.Min, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	status := 0
	for _, ip := range cmd.Args() {
		if err := cli.client.FipRelease(ip); err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			status = 1
			continue
		}
	}
	if status != 0 {
		return Cli.StatusError{StatusCode: status}
	}
	return nil
}

// CmdFipAssociate connects a container to a floating IP
//
// Usage: docker fip associate [OPTIONS] <FIP> <CONTAINER>
func (cli *DockerCli) CmdFipAssociate(args ...string) error {
	cmd := Cli.Subcmd("fip associate", []string{"FIP CONTAINER"}, "Connects a container to a floating IP", false)
	cmd.Require(flag.Min, 2)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}
	return cli.client.FipAssociate(cmd.Arg(0), cmd.Arg(1))
}

// CmdFipDeassociate disconnects a container from a floating IP
//
// Usage: docker fip deassociate <CONTAINER>
func (cli *DockerCli) CmdFipDeassociate(args ...string) error {
	cmd := Cli.Subcmd("fip deassociate", []string{"CONTAINER"}, "Disconnects container from a floating IP", false)
	//force := cmd.Bool([]string{"f", "-force"}, false, "Force the container to disconnect from a floating IP")
	cmd.Require(flag.Exact, 1)
	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	ip, err := cli.client.FipDeassociate(cmd.Arg(0))
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.out, "%s\n", ip)
	return nil
}

// CmdFipLs lists all the fips
//
// Usage: docker fip ls [OPTIONS]
func (cli *DockerCli) CmdFipLs(args ...string) error {
	cmd := Cli.Subcmd("fip ls", nil, "Lists fips", true)

	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")

	cmd.Require(flag.Exact, 0)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process after get response from server.
	fipFilterArgs := filters.NewArgs()
	for _, f := range flFilter.GetAll() {
		if fipFilterArgs, err = filters.ParseFlag(f, fipFilterArgs); err != nil {
			return err
		}
	}

	options := types.NetworkListOptions{
		Filters: fipFilterArgs,
	}

	fips, err := cli.client.FipList(options)
	if err != nil {
		return err
	}
	for _, ip := range fips {
		fmt.Fprintf(cli.out, "%s\n", ip)
	}

	return nil
}

func fipUsage() string {
	fipCommands := map[string]string{
		"allocate":    "Allocate a or some IPs",
		"associate":   "Associate floating IP to container",
		"deassociate": "Deassociate floating IP from conainer",
		"ls":          "List all floating IPs",
		"release":     "Release a floating IP",
	}

	help := "Commands:\n"

	for cmd, description := range fipCommands {
		help += fmt.Sprintf("  %-25.25s%s\n", cmd, description)
	}

	help += fmt.Sprintf("\nRun 'docker fip COMMAND --help' for more information on a command.")
	return help
}
