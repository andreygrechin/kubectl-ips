package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/yaml"
)

var (
	// ErrExpectedTable is returned when the object is not a metav1.Table.
	ErrExpectedTable = errors.New("expected metav1.Table")
	// ErrExpectedPodList is returned when the object is not a PodList.
	ErrExpectedPodList = errors.New("expected PodList")
)

// ResourcePrinter is an interface for printing Kubernetes objects.
type ResourcePrinter interface {
	PrintObj(obj runtime.Object, out io.Writer) error
}

func createPrinter(outputFormat string, noHeaders, showNamespace bool) (ResourcePrinter, error) {
	switch outputFormat {
	case jsonFormat:
		return &jsonPrinter{}, nil
	case yamlFormat:
		return &yamlPrinter{}, nil
	case nameFormat:
		return &namePrinter{showNamespace: showNamespace}, nil
	case tableFormat, wideFormat, "":
		options := printers.PrintOptions{
			NoHeaders: noHeaders,
		}

		return printers.NewTablePrinter(options), nil
	default:
		return nil, ErrUnsupportedFormat
	}
}

type jsonPrinter struct{}

func (p *jsonPrinter) PrintObj(obj runtime.Object, out io.Writer) error {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if _, err = fmt.Fprintln(out, string(data)); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

type yamlPrinter struct{}

func (p *yamlPrinter) PrintObj(obj runtime.Object, out io.Writer) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	if _, err = fmt.Fprint(out, string(data)); err != nil {
		return fmt.Errorf("failed to write YAML: %w", err)
	}

	return nil
}

type namePrinter struct {
	showNamespace bool
}

func (p *namePrinter) PrintObj(obj runtime.Object, out io.Writer) error {
	table, ok := obj.(*metav1.Table)
	if !ok {
		return ErrExpectedTable
	}

	for _, row := range table.Rows {
		const minCellsForName = 2
		if len(row.Cells) < minCellsForName {
			continue
		}
		if p.showNamespace && len(row.Cells) > 0 {
			_, _ = fmt.Fprintf(out, "%s/%s\n", row.Cells[0], row.Cells[1])
		} else if len(row.Cells) > 0 {
			_, _ = fmt.Fprintf(out, "%s\n", row.Cells[0])
		}
	}

	return nil
}

type ipOnlyPrinter struct{}

func (p *ipOnlyPrinter) PrintObj(obj runtime.Object, out io.Writer) error {
	pods, ok := obj.(*corev1.PodList)
	if !ok {
		return ErrExpectedPodList
	}

	podIPs := extractPodIPsWithPods(pods)
	sortPodIPsWithPods(podIPs)

	for _, item := range podIPs {
		_, _ = fmt.Fprintf(out, "%s\n", item.ip)
	}

	return nil
}
