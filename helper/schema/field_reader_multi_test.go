package schema

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestMultiLevelFieldReaderReadFieldExact(t *testing.T) {
	cases := map[string]struct {
		Addr    []string
		Schema  *Schema
		Readers []FieldReader
		Level   string
		Result  FieldReadResult
	}{
		"specific": {
			Addr: []string{"foo"},

			Schema: &Schema{
				Type: TypeString,
			},

			Readers: []FieldReader{
				&MapFieldReader{
					Map: map[string]string{
						"foo": "bar",
					},
				},
				&MapFieldReader{
					Map: map[string]string{
						"foo": "baz",
					},
				},
				&MapFieldReader{
					Map: map[string]string{},
				},
			},

			Level: "1",
			Result: FieldReadResult{
				Value:  "baz",
				Exists: true,
			},
		},
	}

	for name, tc := range cases {
		readers := make(map[string]FieldReader)
		levels := make([]string, len(tc.Readers))
		for i, r := range tc.Readers {
			is := strconv.FormatInt(int64(i), 10)
			readers[is] = r
			levels[i] = is
		}

		r := &MultiLevelFieldReader{
			Readers: readers,
			Levels:  levels,
		}

		out, err := r.ReadFieldExact(tc.Addr, tc.Schema, tc.Level)
		if err != nil {
			t.Fatalf("%s: err: %s", name, err)
		}

		if !reflect.DeepEqual(tc.Result, out) {
			t.Fatalf("%s: bad: %#v", name, out)
		}
	}
}

func TestMultiLevelFieldReaderReadFieldMerge(t *testing.T) {
	cases := map[string]struct {
		Addr    []string
		Schema  *Schema
		Readers []FieldReader
		Result  FieldReadResult
	}{
		"stringInDiff": {
			Addr: []string{"availability_zone"},

			Schema: &Schema{
				Type:     TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			Readers: []FieldReader{
				&DiffFieldReader{
					Source: &MapFieldReader{
						Map: map[string]string{
							"availability_zone": "foo",
						},
					},

					Diff: &terraform.InstanceDiff{
						Attributes: map[string]*terraform.ResourceAttrDiff{
							"availability_zone": &terraform.ResourceAttrDiff{
								Old:         "foo",
								New:         "bar",
								RequiresNew: true,
							},
						},
					},
				},
			},

			Result: FieldReadResult{
				Value:  "bar",
				Exists: true,
			},
		},

		"lastLevelComputed": {
			Addr: []string{"availability_zone"},

			Schema: &Schema{
				Type:     TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			Readers: []FieldReader{
				&MapFieldReader{
					Map: map[string]string{
						"availability_zone": "foo",
					},
				},

				&DiffFieldReader{
					Source: &MapFieldReader{
						Map: map[string]string{
							"availability_zone": "foo",
						},
					},

					Diff: &terraform.InstanceDiff{
						Attributes: map[string]*terraform.ResourceAttrDiff{
							"availability_zone": &terraform.ResourceAttrDiff{
								Old:         "foo",
								New:         "bar",
								NewComputed: true,
							},
						},
					},
				},
			},

			Result: FieldReadResult{
				Value:    "",
				Exists:   true,
				Computed: true,
			},
		},

		"list of maps with removal in diff": {
			Addr: []string{"config_vars"},

			Schema: &Schema{
				Type:     TypeList,
				Optional: true,
				Computed: true,
				Elem: &Schema{
					Type: TypeMap,
				},
			},

			Readers: []FieldReader{
				&DiffFieldReader{
					Source: &MapFieldReader{
						Map: map[string]string{
							"config_vars.#":     "2",
							"config_vars.0.foo": "bar",
							"config_vars.0.bar": "bar",
							"config_vars.1.bar": "baz",
						},
					},

					Diff: &terraform.InstanceDiff{
						Attributes: map[string]*terraform.ResourceAttrDiff{
							"config_vars.0.bar": &terraform.ResourceAttrDiff{
								NewRemoved: true,
							},
						},
					},
				},
			},

			Result: FieldReadResult{
				Value: []interface{}{
					map[string]interface{}{
						"foo": "bar",
					},
					map[string]interface{}{
						"bar": "baz",
					},
				},
				Exists: true,
			},
		},

		"first level only": {
			Addr: []string{"foo"},

			Schema: &Schema{
				Type: TypeString,
			},

			Readers: []FieldReader{
				&MapFieldReader{
					Map: map[string]string{
						"foo": "bar",
					},
				},
				&MapFieldReader{
					Map: map[string]string{},
				},
			},

			Result: FieldReadResult{
				Value:  "bar",
				Exists: true,
			},
		},
	}

	for name, tc := range cases {
		readers := make(map[string]FieldReader)
		levels := make([]string, len(tc.Readers))
		for i, r := range tc.Readers {
			is := strconv.FormatInt(int64(i), 10)
			readers[is] = r
			levels[i] = is
		}

		r := &MultiLevelFieldReader{
			Readers: readers,
			Levels:  levels,
		}

		out, err := r.ReadFieldMerge(tc.Addr, tc.Schema, levels[len(levels)-1])
		if err != nil {
			t.Fatalf("%s: err: %s", name, err)
		}

		if !reflect.DeepEqual(tc.Result, out) {
			t.Fatalf("%s: bad: %#v", name, out)
		}
	}
}
