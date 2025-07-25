package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
)

const (
	jsonFormat  = "json"
	yamlFormat  = "yaml"
	nameFormat  = "name"
	tableFormat = "table"
	wideFormat  = "wide"
)

var ipsExample = `
  # list all pod IP addresses in the current namespace
  %[1]s ips

  # list all pod IP addresses in all namespaces
  %[1]s ips --all-namespaces

  # list pod IP addresses in a specific namespace
  %[1]s ips --namespace=kube-system

  # filter pods by label selector
  %[1]s ips --selector=app=nginx

  # show only IP addresses without pod names
  %[1]s ips --show-ips-only

  # show wide output with additional columns
  %[1]s ips -o wide

  # output in JSON format
  %[1]s ips -o json

  # show labels as additional column
  %[1]s ips --show-labels
`

// IPsOptions provides information required to list pod IP addresses.
type IPsOptions struct {
	genericiooptions.IOStreams

	configFlags *genericclioptions.ConfigFlags

	allNamespaces bool
	labelSelector string
	showIPsOnly   bool
	namespace     string
	outputFormat  string
	noHeaders     bool
	showLabels    bool
}

// NewIPsOptions provides an instance of IPsOptions with default values.
func NewIPsOptions(streams genericiooptions.IOStreams) *IPsOptions {
	return &IPsOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

var ErrUnsupportedFormat = errors.New("unsupported output format")

// NewCmdIPs provides a cobra command wrapping IPsOptions.
func NewCmdIPs(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewIPsOptions(streams)

	cmd := &cobra.Command{
		Use:          "ips [flags]",
		Short:        "List IP addresses from Kubernetes pods",
		Example:      fmt.Sprintf(ipsExample, "kubectl"),
		SilenceUsage: true,
		Annotations: map[string]string{
			cobra.CommandDisplayNameAnnotation: "kubectl ips",
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", false,
		"If true, list IP addresses from pods in all namespaces")
	cmd.Flags().StringVarP(&o.labelSelector, "selector", "l", "",
		"Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&o.showIPsOnly, "show-ips-only", false, "If true, show only IP addresses without pod names")
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", "table",
		"Output format. One of: (table, wide, json, yaml, name)")
	cmd.Flags().BoolVar(&o.noHeaders, "no-headers", false,
		"When using the default or custom output format, don't print headers")
	cmd.Flags().BoolVar(&o.showLabels, "show-labels", false, "When printing, show all labels as the last column")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all information required for listing pod IPs.
func (o *IPsOptions) Complete(cmd *cobra.Command, _ []string) error {
	var err error
	o.namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("failed to get namespace flag: %w", err)
	}

	if o.allNamespaces {
		o.namespace = ""
	}

	if o.namespace == "" && !o.allNamespaces {
		if o.configFlags.Namespace != nil && *o.configFlags.Namespace != "" {
			o.namespace = *o.configFlags.Namespace
		} else {
			o.namespace = "default"
		}
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided.
func (o *IPsOptions) Validate() error {
	switch o.outputFormat {
	case tableFormat, wideFormat, jsonFormat, yamlFormat, nameFormat, "":
		// valid formats
	default:
		return ErrUnsupportedFormat
	}

	return nil
}

// SetOutputFormat sets the output format for testing purposes.
func (o *IPsOptions) SetOutputFormat(format string) {
	o.outputFormat = format
}

// Run lists IP addresses from pods based on the provided options.
func (o *IPsOptions) Run() error {
	pods, err := o.getPods()
	if err != nil {
		return err
	}

	// Handle legacy --show-ips-only flag
	if o.showIPsOnly {
		printer := &ipOnlyPrinter{}

		return printer.PrintObj(pods, o.Out)
	}

	// Check if we have any pods
	if len(pods.Items) == 0 {
		return o.printNoPodsFound()
	}

	// Generate table for new output formats
	wide := o.outputFormat == wideFormat
	table := generateTable(pods, o.allNamespaces, wide, o.showLabels)

	if len(table.Rows) == 0 {
		return o.printNoPodsFound()
	}

	// Create and use appropriate printer
	printer, err := createPrinter(o.outputFormat, o.noHeaders, o.allNamespaces)
	if err != nil {
		return err
	}

	if err := printer.PrintObj(table, o.Out); err != nil {
		return fmt.Errorf("failed to print object: %w", err)
	}

	return nil
}

func (o *IPsOptions) getPods() (*corev1.PodList, error) {
	config, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()
	listOptions := metav1.ListOptions{}
	if o.labelSelector != "" {
		listOptions.LabelSelector = o.labelSelector
	}

	var pods *corev1.PodList
	if o.allNamespaces {
		pods, err = clientset.CoreV1().Pods("").List(ctx, listOptions)
	} else {
		pods, err = clientset.CoreV1().Pods(o.namespace).List(ctx, listOptions)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods, nil
}

func (o *IPsOptions) printNoPodsFound() error {
	namespace := o.namespace
	if namespace == "" {
		namespace = "all namespaces"
	}
	selectorInfo := ""
	if o.labelSelector != "" {
		selector, _ := labels.Parse(o.labelSelector)
		selectorInfo = fmt.Sprintf(" matching selector %q", selector.String())
	}
	_, _ = fmt.Fprintf(o.Out, "No pods found in %s%s\n", namespace, selectorInfo)

	return nil
}
