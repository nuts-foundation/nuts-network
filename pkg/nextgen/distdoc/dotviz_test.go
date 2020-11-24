package distdoc

import (
	"fmt"
	"strings"
)

// DotGraphVisitor is a graph visitor that outputs the walked path as "dot" diagram.
type DotGraphVisitor struct {
	graph       *memoryDAG
	output      string
	aliases     map[string]string
	counter     int
	labelStyle  LabelStyle
	showAliases bool
	showContent bool

	nodes []string
	edges []string
}

// LabelStyle defines node label styles for DotGraphVisitor.
type LabelStyle int

const (
	// ShowAliasLabelStyle is a style that uses integer aliases for node labels.
	ShowAliasLabelStyle LabelStyle = iota
	// ShowAliasLabelStyle is a style that uses the references of nodes as label.
	ShowRefLabelStyle LabelStyle = iota
)

func NewDotGraphVisitor(graph *memoryDAG, labelStyle LabelStyle) *DotGraphVisitor {
	return &DotGraphVisitor{
		graph:      graph,
		aliases:    map[string]string{},
		labelStyle: labelStyle,
	}
}

func (d *DotGraphVisitor) Accept(document Document) {
	d.counter++
	d.nodes = append(d.nodes, fmt.Sprintf("  \"%s\"[label=\"%s (%d)\"]", document.Ref().String(), d.label(document), d.counter))
	for _, prev := range document.Previous() {
		d.edges = append(d.edges, fmt.Sprintf("  \"%s\" -> \"%s\"", prev.String(), document.Ref().String()))
	}
}

func (d *DotGraphVisitor) Render() string {
	var lines []string
	lines = append(lines, "digraph {")
	lines = append(lines, d.nodes...)
	lines = append(lines, d.edges...)
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func (d *DotGraphVisitor) label(document Document) string {
	switch d.labelStyle {
	case ShowAliasLabelStyle:
		if alias, ok := d.aliases[document.Ref().String()]; ok {
			return alias
		} else {
			alias = fmt.Sprintf("%d", len(d.aliases)+1)
			d.aliases[document.Ref().String()] = alias
			return alias
		}
	case ShowRefLabelStyle:
		fallthrough
	default:
		return document.Ref().String()
	}
}
