package render_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/oche/diagram/internal/iac/graph"
	_ "github.com/oche/diagram/internal/iac/graph/props"
	"github.com/oche/diagram/internal/iac/render"
)

var update = flag.Bool("update", false, "rewrite *.golden files from current render output")

func TestRender_Golden(t *testing.T) {
	cases, err := filepath.Glob("../testdata/stacks/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range cases {
		name := filepath.Base(dir)
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, "graph.golden.json"))
			if err != nil {
				t.Fatal(err)
			}
			var g graph.ResourceGraph
			if err := json.Unmarshal(raw, &g); err != nil {
				t.Fatalf("parse graph: %v", err)
			}

			tf, err := render.RenderTerraform(&g)
			if err != nil {
				t.Fatalf("terraform render: %v", err)
			}
			compareGolden(t, filepath.Join(dir, "main.tf.golden"), tf)

			cfn, err := render.RenderCloudFormation(&g)
			if err != nil {
				t.Fatalf("cfn render: %v", err)
			}
			compareGolden(t, filepath.Join(dir, "template.cfn.yaml.golden"), cfn)
		})
	}
}

func compareGolden(t *testing.T, path, got string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v (run with -update to create)", path, err)
	}
	if string(want) != got {
		t.Errorf("%s: golden mismatch.\nFirst diff (expect left, got right):\n%s", path, firstDiff(string(want), got))
	}
}

func firstDiff(want, got string) string {
	max := len(want)
	if len(got) < max {
		max = len(got)
	}
	for i := 0; i < max; i++ {
		if want[i] != got[i] {
			start := i - 30
			if start < 0 {
				start = 0
			}
			endWant := i + 30
			if endWant > len(want) {
				endWant = len(want)
			}
			endGot := i + 30
			if endGot > len(got) {
				endGot = len(got)
			}
			return "want: " + want[start:endWant] + "\ngot:  " + got[start:endGot]
		}
	}
	if len(want) != len(got) {
		return "length mismatch"
	}
	return "no diff"
}
