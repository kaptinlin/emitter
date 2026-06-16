package emitter

import (
	"os"
	"strings"
	"testing"
)

func TestMaintenanceTasksExposeRequiredGates(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatal(err)
	}

	text := string(data)
	for _, task := range []string{"deps:update", "deps:submodules", "lint", "test"} {
		if !strings.Contains(text, "\n  "+task+":\n") {
			t.Fatalf("Taskfile.yml missing %q task", task)
		}
	}
}

func TestReadmeGoVersionMatchesModule(t *testing.T) {
	t.Parallel()

	mod, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}

	goVersion := ""
	for _, line := range strings.Split(string(mod), "\n") {
		if version, ok := strings.CutPrefix(line, "go "); ok {
			goVersion = strings.TrimSpace(version)
			break
		}
	}
	if goVersion == "" {
		t.Fatal("go.mod missing go directive")
	}

	want := "Requires Go " + goVersion + " or newer."
	if !strings.Contains(string(readme), want) {
		t.Fatalf("README.md missing %q", want)
	}
}
