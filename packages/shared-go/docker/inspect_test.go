package docker

import (
	"encoding/json"
	"os"
	"testing"
)

func TestContainerInspectDecode(t *testing.T) {
	raw, err := os.ReadFile("inspect_sample.json")
	if err != nil {
		t.Fatal("got unexpected error reading inspect_sample.json", err)
	}

	var inspect_output []ContainerInspect
	if err := json.Unmarshal(raw, &inspect_output); err != nil {
		t.Fatal("got unexpected error decoding inspect sample", err)
	}

	if len(inspect_output) != 1 {
		t.Fatalf("got %d results, want 1", len(inspect_output))
	}

	container := inspect_output[0]

	if want := "/mrbovino-app"; container.Name != want {
		t.Errorf("got name '%s', want '%s'", container.Name, want)
	}
	if want := "exited"; container.State.Status != want {
		t.Errorf("got state status '%s', want '%s'", container.State.Status, want)
	}
	if want := "capybaracloud/mrbovino:123"; container.Config.Image != want {
		t.Errorf("got config image '%s', want '%s'", container.Config.Image, want)
	}
	if container.NetworkSettings.Ports == nil {
		t.Error("got nil NetworkSettings.Ports, want non-nil map")
	}
	if _, ok := container.NetworkSettings.Networks["bridge"]; !ok {
		t.Error("got missing 'bridge' network, want present")
	}
}
