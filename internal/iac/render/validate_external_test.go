//go:build iac_external
// +build iac_external

package render_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_External(t *testing.T) {
	cases, _ := filepath.Glob("../testdata/stacks/*")
	for _, dir := range cases {
		name := filepath.Base(dir)
		t.Run("Terraform/"+name, func(t *testing.T) {
			work := t.TempDir()
			tfPath := filepath.Join(work, "main.tf")
			mustCopy(t, filepath.Join(dir, "main.tf.golden"), tfPath)
			env := terraformEnv(t, work, tfPath)
			run(t, work, env, "terraform", "init", "-backend=false", "-input=false", "-no-color")
			run(t, work, env, "terraform", "validate", "-no-color")
			args := []string{"plan", "-refresh=false", "-input=false", "-no-color", "-lock=false"}
			if needsDBPassword(t, tfPath) {
				args = append(args, "-var", "db_password=changeme123")
			}
			run(t, work, env, "terraform", args...)
		})
		t.Run("CFN/"+name, func(t *testing.T) {
			work := t.TempDir()
			dst := filepath.Join(work, "template.yaml")
			mustCopy(t, filepath.Join(dir, "template.cfn.yaml.golden"), dst)
			run(t, work, nil, "cfn-lint", dst)
			run(t, work, nil, "aws", "cloudformation", "validate-template", "--template-body", "file://"+dst, "--no-cli-pager")
		})
	}
}

func run(t *testing.T, dir string, extraEnv []string, cmd string, args ...string) {
	t.Helper()
	c := exec.Command(cmd, args...)
	c.Dir = dir
	c.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
		"AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"AWS_DEFAULT_REGION=us-east-1",
	)
	c.Env = append(c.Env, extraEnv...)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", cmd, args, err, out)
	}
}

func mustCopy(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func needsDBPassword(t *testing.T, path string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(b), `variable "db_password"`)
}

func terraformEnv(t *testing.T, work, tfPath string) []string {
	t.Helper()

	var env []string
	if strings.Contains(mustRead(t, tfPath), `"lambda.zip"`) {
		lambdaZip := filepath.Join(work, "lambda.zip")
		if err := os.WriteFile(lambdaZip, []byte("placeholder lambda payload"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mirror := os.Getenv("IAC_TERRAFORM_PLUGIN_MIRROR")
	if mirror == "" {
		const defaultMirror = "/tmp/iac-tools/plugin-cache"
		if _, err := os.Stat(defaultMirror); err == nil {
			mirror = defaultMirror
		}
	}
	if mirror == "" {
		return env
	}

	rcPath := filepath.Join(work, "terraform.rc")
	rcBody := strings.Join([]string{
		"provider_installation {",
		"  filesystem_mirror {",
		`    path    = "` + mirror + `"`,
		"  }",
		"  direct {",
		`    exclude = ["registry.terraform.io/hashicorp/aws"]`,
		"  }",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(rcPath, []byte(rcBody), 0o644); err != nil {
		t.Fatal(err)
	}
	env = append(env, "TF_CLI_CONFIG_FILE="+rcPath)
	return env
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
