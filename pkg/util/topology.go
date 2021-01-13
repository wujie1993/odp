package util

import (
	"bytes"
	"strings"
	"text/template"
)

const (
	TopologyFormatCSV     = "csv"
	TopologyFormatMermaid = "mermaid"

	MermaidFlowOrientationTB = "TB"
	MermaidFlowOrientationLR = "LR"

	MermaidFlowShapeRectangle        = "rectangle"
	MermaidFlowShapeRoundEdges       = "roundEdges"
	MermaidFlowShapeStadium          = "stadium"
	MermaidFlowShapeSubroutine       = "subroutine"
	MermaidFlowShapeCylindrical      = "cylindrical"
	MermaidFlowShapeCircle           = "circle"
	MermaidFlowShapeAsymmetric       = "asymmetric"
	MermaidFlowShapeRhombus          = "rhombus"
	MermaidFlowShapeHexagon          = "hexagon"
	MermaidFlowShapeParallelogram    = "parallelogram"
	MermaidFlowShapeParallelogramAlt = "parallelogramAlt"
	MermaidFlowShapeTrapezoid        = "trapezoid"
	MermaidFlowShapeTrapezoidAlt     = "trapezoidAlt"

	MermaidFlowLinkTypeNormal = "normal"
	MermaidFlowLinkTypeThick  = "thick"
	MermaidFlowLinkTypeDotted = "dotted"

	MermaidFlowArrowRound    = "round"
	MermaidFlowArrowTriangle = "triangle"
	MermaidFlowArrowFork     = "fork"
)

const (
	mermaidTpl = `
{{- define "node" }}
{{- if eq .Shape "rectangle" }}[{{ end -}}
{{- if eq .Shape "roundEdges" }}({{ end -}}
{{- if eq .Shape "stadium" }}([{{ end -}}
{{- if eq .Shape "subroutine" }}[[{{ end -}}
{{- if eq .Shape "cylindrical" }}[({{ end -}}
{{- if eq .Shape "circle" }}(({{ end -}}
{{- if eq .Shape "asymmetric" }}>{{ end -}}
{{- if eq .Shape "rhombus" }}{{ print "{" }}{{ end -}}
{{- if eq .Shape "hexagon" }}{{ print "{{" }}{{ end -}}
{{- if eq .Shape "parallelogram" }}[/{{ end -}}
{{- if eq .Shape "parallelogramAlt" }}[\{{ end -}}
{{- if eq .Shape "trapezoid" }}[/{{ end -}}
{{- if eq .Shape "trapezoidAlt" }}[\{{ end -}}
"{{ .Desc }}"
{{- if eq .Shape "rectangle" }}]{{ end }}
{{- if eq .Shape "roundEdges" }}){{ end }}
{{- if eq .Shape "stadium" }}]){{ end }}
{{- if eq .Shape "subroutine" }}]]{{ end }}
{{- if eq .Shape "cylindrical" }})]{{ end }}
{{- if eq .Shape "circle" }})){{ end }}
{{- if eq .Shape "asymmetric" }}]{{ end }}
{{- if eq .Shape "rhombus" }}{{ print "}" }}{{ end -}}
{{- if eq .Shape "hexagon" }}{{ print "}}" }}{{ end -}}
{{- if eq .Shape "parallelogram" }}/]{{ end -}}
{{- if eq .Shape "parallelogramAlt" }}\]{{ end -}}
{{- if eq .Shape "trapezoid" }}\]{{ end -}}
{{- if eq .Shape "trapezoidAlt" }}/]{{ end -}}
{{- end }}
{{- define "link" }}
	{{- .Src.Name }} {{ if .Src.Arrow }}
		{{- if eq .Src.Arrow "round" }}o{{ end }}
		{{- if eq .Src.Arrow "triangle" }}<{{ end }}
		{{- if eq .Src.Arrow "fork" }}x{{ end }}
	{{- end }}

	{{- if eq .LinkType "normal" "dotted" }}-
	{{- else if eq .LinkType "thick" }}=
	{{- else }}-
	{{- end }}

	{{- if eq .LinkType "normal" }}-
	{{- else if eq .LinkType "dotted" }}.
	{{- else if eq .LinkType "thick" }}=
	{{- else }}-
	{{- end }}

	{{- if eq .LinkType "normal" "dotted" }}-
	{{- else if eq .LinkType "thick" }}=
	{{- else }}-
	{{- end }}

	{{- if .Dst.Arrow }}
		{{- if eq .Dst.Arrow "round" }}o{{ end }}
		{{- if eq .Dst.Arrow "triangle" }}>{{ end }}
		{{- if eq .Dst.Arrow "fork" }}x{{ end }}
	{{- end }}

	{{- if .LinkText }}|{{ .LinkText }}|{{ end }} {{ .Dst.Name }}
{{- end }}
{{- define "nodes" }}
	{{- range $name,$node := .Nodes }}
	{{ $name }}{{ template "node" $node }}
	{{- end }}
{{- end }}
{{- define "links" }}
	{{- range .Links }}
	{{ template "link" . }}
	{{- end }}
{{- end}}
flowchart {{ .Orientation }}
	{{- template "nodes" . }}
	{{- template "links" . }}
	{{- range $name, $subGraph := .SubGraphs }}
	subgraph {{ $name }}
		{{- template "nodes" $subGraph }}
		{{- template "links" $subGraph }}
	end
	{{- end }}
`
)

type CSVWriter struct {
	buffer bytes.Buffer
}

func (w *CSVWriter) WriteRow(cells []string) error {
	for index, cell := range cells {
		cells[index] = "\"" + cell + "\""
	}
	_, err := w.buffer.WriteString(strings.Join(cells, ",") + "\n")
	return err
}

func (w CSVWriter) String() string {
	return w.buffer.String()
}

func NewCSVWriter(headers []string) CSVWriter {
	w := CSVWriter{}
	// 写入UTF-8 BOM，避免使用Microsoft Excel打开乱码
	w.buffer.WriteString("\xEF\xBB\xBF")
	// 写入列头
	w.buffer.WriteString(strings.Join(headers, ",") + "\n")
	return w
}

type MermaidFlow struct {
	Orientation string
	Nodes       map[string]MermaidFlowNode
	Links       []MermaidFlowLink
	SubGraphs   map[string]MermaidFlowSubGraph
}

type MermaidFlowSubGraph struct {
	Nodes map[string]MermaidFlowNode
	Links []MermaidFlowLink
}

type MermaidFlowNode struct {
	Desc  string
	Shape string
}

type MermaidFlowLink struct {
	Src      MermaidFlowLinkEndpoint
	Dst      MermaidFlowLinkEndpoint
	LinkType string
	LinkText string
}

type MermaidFlowLinkEndpoint struct {
	Name  string
	Arrow string
}

func RenderMermaidFlowChart(obj MermaidFlow) (string, error) {
	tpl, err := template.New("mermaidFlow").Parse(mermaidTpl)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, obj); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
