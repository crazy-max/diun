package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeNode(t *testing.T) {
	testCases := []struct {
		desc     string
		node     *Node
		expected map[string]string
	}{
		{
			desc: "1 label",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "aaa", Value: "bar"},
				},
			},
			expected: map[string]string{
				"traefik.aaa": "bar",
			},
		},
		{
			desc: "2 labels",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "aaa", Value: "bar"},
					{Name: "bbb", Value: "bur"},
				},
			},
			expected: map[string]string{
				"traefik.aaa": "bar",
				"traefik.bbb": "bur",
			},
		},
		{
			desc: "2 labels, 1 disabled",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "aaa", Value: "bar"},
					{Name: "bbb", Value: "bur", Disabled: true},
				},
			},
			expected: map[string]string{
				"traefik.aaa": "bar",
			},
		},
		{
			desc: "2 levels",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "aaa", Value: "bar"},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo.aaa": "bar",
			},
		},
		{
			desc: "3 levels",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
							{Name: "aaa", Value: "bar"},
						}},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo.bar.aaa": "bar",
			},
		},
		{
			desc: "2 levels, same root",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
							{Name: "aaa", Value: "bar"},
							{Name: "bbb", Value: "bur"},
						}},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo.bar.aaa": "bar",
				"traefik.foo.bar.bbb": "bur",
			},
		},
		{
			desc: "several levels, different root",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "bar", Children: []*Node{
						{Name: "ccc", Value: "bir"},
					}},
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
							{Name: "aaa", Value: "bar"},
						}},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo.bar.aaa": "bar",
				"traefik.bar.ccc":     "bir",
			},
		},
		{
			desc: "multiple labels, multiple levels",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "bar", Children: []*Node{
						{Name: "ccc", Value: "bir"},
					}},
					{Name: "foo", Children: []*Node{
						{Name: "bar", Children: []*Node{
							{Name: "aaa", Value: "bar"},
							{Name: "bbb", Value: "bur"},
						}},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo.bar.aaa": "bar",
				"traefik.foo.bar.bbb": "bur",
				"traefik.bar.ccc":     "bir",
			},
		},
		{
			desc: "slice of struct syntax",
			node: &Node{
				Name: "traefik",
				Children: []*Node{
					{Name: "foo", Children: []*Node{
						{Name: "[0]", Children: []*Node{
							{Name: "aaa", Value: "bar0"},
							{Name: "bbb", Value: "bur0"},
						}},
						{Name: "[1]", Children: []*Node{
							{Name: "aaa", Value: "bar1"},
							{Name: "bbb", Value: "bur1"},
						}},
					}},
				},
			},
			expected: map[string]string{
				"traefik.foo[0].aaa": "bar0",
				"traefik.foo[0].bbb": "bur0",
				"traefik.foo[1].aaa": "bar1",
				"traefik.foo[1].bbb": "bur1",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			labels := EncodeNode(test.node)

			assert.Equal(t, test.expected, labels)
		})
	}
}
