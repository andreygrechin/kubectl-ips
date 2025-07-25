package cmd_test

import (
	"bytes"
	"testing"

	"github.com/andreygrechin/kubectl-ips/pkg/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewIPsOptions(t *testing.T) {
	streams := genericiooptions.NewTestIOStreamsDiscard()
	options := cmd.NewIPsOptions(streams)

	assert.NotNil(t, options)
	assert.Equal(t, streams, options.IOStreams)
}

func TestNewCmdIPs(t *testing.T) {
	streams := genericiooptions.NewTestIOStreamsDiscard()
	command := cmd.NewCmdIPs(streams)

	assert.NotNil(t, command)
	assert.Equal(t, "ips [flags]", command.Use)
	assert.Contains(t, command.Short, "List IP addresses from Kubernetes pods")
}

func TestIPsOptionsComplete(t *testing.T) {
	tests := map[string]struct {
		args          []string
		flagNamespace string
		flagAllNS     bool
		expectedNS    string
		expectedAllNS bool
		expectError   bool
	}{
		"default namespace": {
			args:          []string{},
			expectedNS:    "default",
			expectedAllNS: false,
		},
		"specific namespace": {
			args:          []string{},
			flagNamespace: "kube-system",
			expectedNS:    "kube-system",
			expectedAllNS: false,
		},
		"all namespaces": {
			args:          []string{},
			flagAllNS:     true,
			expectedNS:    "",
			expectedAllNS: true,
		},
		"all namespaces overrides specific": {
			args:          []string{},
			flagNamespace: "kube-system",
			flagAllNS:     true,
			expectedNS:    "",
			expectedAllNS: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			streams := genericiooptions.NewTestIOStreamsDiscard()
			options := cmd.NewIPsOptions(streams)
			cmdObj := cmd.NewCmdIPs(streams)

			if tc.flagNamespace != "" {
				require.NoError(t, cmdObj.Flags().Set("namespace", tc.flagNamespace))
			}
			if tc.flagAllNS {
				require.NoError(t, cmdObj.Flags().Set("all-namespaces", "true"))
			}

			err := options.Complete(cmdObj, tc.args)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPsOptionsValidate(t *testing.T) {
	tests := map[string]struct {
		outputFormat string
		expectError  bool
	}{
		"valid table format": {
			outputFormat: "table",
			expectError:  false,
		},
		"valid wide format": {
			outputFormat: "wide",
			expectError:  false,
		},
		"valid json format": {
			outputFormat: "json",
			expectError:  false,
		},
		"valid yaml format": {
			outputFormat: "yaml",
			expectError:  false,
		},
		"valid name format": {
			outputFormat: "name",
			expectError:  false,
		},
		"empty format": {
			outputFormat: "",
			expectError:  false,
		},
		"invalid format": {
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			streams := genericiooptions.NewTestIOStreamsDiscard()
			options := cmd.NewIPsOptions(streams)
			options.SetOutputFormat(tc.outputFormat)

			err := options.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIPsCommandFlags(t *testing.T) {
	streams := genericiooptions.NewTestIOStreamsDiscard()
	command := cmd.NewCmdIPs(streams)

	flags := []string{
		"all-namespaces",
		"selector",
		"show-ips-only",
		"namespace",
		"kubeconfig",
		"context",
		"output",
		"no-headers",
		"show-labels",
	}

	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			f := command.Flags().Lookup(flag)
			assert.NotNil(t, f, "flag %s should exist", flag)
		})
	}

	shortA := command.Flags().ShorthandLookup("A")
	assert.NotNil(t, shortA)
	assert.Equal(t, "all-namespaces", shortA.Name)

	shortL := command.Flags().ShorthandLookup("l")
	assert.NotNil(t, shortL)
	assert.Equal(t, "selector", shortL.Name)

	shortO := command.Flags().ShorthandLookup("o")
	assert.NotNil(t, shortO)
	assert.Equal(t, "output", shortO.Name)
}

func TestIPsCommandExecution(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	streams := genericiooptions.IOStreams{
		In:     nil,
		Out:    &out,
		ErrOut: &errOut,
	}

	command := cmd.NewCmdIPs(streams)
	command.SetArgs([]string{"--help"})
	command.SetOut(&out)
	command.SetErr(&errOut)

	err := command.Execute()
	require.NoError(t, err)

	helpOutput := out.String() + errOut.String()
	assert.Contains(t, helpOutput, "List IP addresses from Kubernetes pods")
	assert.Contains(t, helpOutput, "--all-namespaces")
	assert.Contains(t, helpOutput, "--selector")
	assert.Contains(t, helpOutput, "--show-ips-only")
}
