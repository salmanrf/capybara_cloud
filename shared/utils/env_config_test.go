package utils

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestConfigLoadGetSet(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	
	tests := []struct{
		env_path string
		keys []string
	}{
		{
			path.Join(pwd, "samples", ".env.test"),
			[]string{"POSTGRES_URI", "REDIS_URI", "JWT_SECRET"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("should load values from file and have keys: %v", tt.keys), func (t *testing.T) {
			cfg := NewConfig()
			cfg.Load(tt.env_path, tt.keys)
			
			err := cfg.Load(tt.env_path, tt.keys)
			if err != nil {
				t.Fatal(err)
			}

			for _, k := range tt.keys {
				if _, ok := cfg.Get(k); !ok {
					t.Errorf("got key %s doesn't exist, want string", k)
				}
			}
		})
	}
}