package yaml

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	testYAML = `---
test:
  me:
    plz: "OK!"
---
test:
  this:
    too: "OK?"
`
	testUnstructedSlice = []unstructured.Unstructured{
		{
			Object: map[string]any{
				"test": map[string]any{
					"me": map[string]any{
						"plz": "OK!",
					},
				},
			},
		},
		{
			Object: map[string]any{
				"test": map[string]any{
					"this": map[string]any{
						"too": "OK?",
					},
				},
			},
		},
	}
)

func TestLoadData(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				r: strings.NewReader(testYAML),
			},
			want: testUnstructedSlice,
		},
		{
			name: "error",
			args: args{
				r: strings.NewReader("this is wrong"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadData(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadData() = %v, want %v", got, tt.want)
			}
		})
	}
}
