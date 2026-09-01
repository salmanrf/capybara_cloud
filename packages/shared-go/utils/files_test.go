package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestEnsureDirExists(t *testing.T) {
	tests := []struct{
		path string
	}{
		{"/tmp/tests/capybara_builds"},
		{"/tmp/a/b/c/d"},
		{"~/12345"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should ensure path %s exists (create if not already)", tt.path), func (t *testing.T) {
			defer func () {
				os.Remove(tt.path)
			}()
			
			got_err := EnsureDirExists(tt.path)
			if got_err != nil {
				t.Fatal(got_err)
			}

			_, err := os.ReadDir(tt.path)
			if err != nil {
				t.Errorf("got error %s, want nil", err)
			}
		})
	}
}