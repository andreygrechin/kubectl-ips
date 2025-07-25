package main

import (
	"os"

	"github.com/andreygrechin/kubectl-ips/pkg/cmd"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-ips", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewCmdIPs(genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
