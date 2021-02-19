package util_test

import (
	"testing"

	"github.com/wujie1993/waves/pkg/util"
)

func TestGetMermaidTopology(t *testing.T) {
	obj := util.MermaidFlow{
		Orientation: util.MermaidFlowOrientationTB,
		Nodes: map[string]util.MermaidFlowNode{
			"host-172.25.21.24": {
				Desc:  "主机 172.25.21.24",
				Shape: util.MermaidFlowShapeRoundEdges,
			},
			"host-172.25.21.25": {
				Desc:  "主机 172.25.21.25",
				Shape: util.MermaidFlowShapeRoundEdges,
			},
		},
		Links: []util.MermaidFlowLink{
			{
				Src: util.MermaidFlowLinkEndpoint{
					Name: "host-172.25.21.24",
				},
				Dst: util.MermaidFlowLinkEndpoint{
					Name:  "host-172.25.21.25",
					Arrow: util.MermaidFlowArrowTriangle,
				},
				LinkText: "yes",
				LinkType: util.MermaidFlowLinkTypeDotted,
			},
		},
		SubGraphs: map[string]util.MermaidFlowSubGraph{
			"ceshi": {
				Nodes: map[string]util.MermaidFlowNode{
					"host-172.25.21.24": {
						Desc:  "主机 172.25.21.24",
						Shape: util.MermaidFlowShapeTrapezoidAlt,
					},
					"host-172.25.21.25": {
						Desc:  "主机 172.25.21.25",
						Shape: util.MermaidFlowShapeRoundEdges,
					},
				},
				Links: []util.MermaidFlowLink{
					{
						Src: util.MermaidFlowLinkEndpoint{
							Name: "host-172.25.21.24",
						},
						Dst: util.MermaidFlowLinkEndpoint{
							Name:  "host-172.25.21.25",
							Arrow: util.MermaidFlowArrowTriangle,
						},
						LinkText: "yes",
						LinkType: util.MermaidFlowLinkTypeDotted,
					},
				},
			},
		},
	}
	ret, err := util.RenderMermaidFlowChart(obj)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}
