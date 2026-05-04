package utils

import (
	"testing"
)

func TestStringSlugify(t *testing.T) {
	t.Run("should remove spaces and symbols", func (t *testing.T) {
		tests := []struct{
			input  string
			output string
		}{
			{ "Handsome Capybara", "handsome_capybara" },
			{ "SoftwareEnjoyer@@@@", "softwareenjoyer" },
			{ "123$$$$%AVERAGE Fan Enjoyer", "123average_fan_enjoyer" },
		}

		for _, tt := range tests {
			t.Run("should slugify", func (t *testing.T) {
				result := Slugify(tt.input)

				got := result
				want := tt.output

				if got != want {
					t.Errorf("got result %s, want %s", got, want)
				}
			})
		}
	})
}