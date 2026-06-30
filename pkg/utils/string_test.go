package utils

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
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

func TestDockerSafeDateString(t *testing.T) {
	tests := []struct{
		tm time.Time 
	}{
		{time.Now()},
		{time.Now().Add(-1 * time.Hour * 100)},
		{time.Now().Add(5 * time.Hour)},
	}

	for _, tt := range tests {
		got_str := DockerSafeDateString(tt.tm)

		if strings.Contains(got_str, " ") {
			t.Errorf("got date string %s containing space", got_str)
		}
		
		if strings.Contains(got_str, ":") {
			t.Errorf("got date string %s containing ':'", got_str)
		}

		year, month, day := tt.tm.Date()
		hour, minute, second := tt.tm.Hour(), tt.tm.Minute(), tt.tm.Second()
		format := fmt.Sprintf(
			"^%s-%02s-%02s-%02s-%02s-%02s$", 
			fmt.Sprintf("%d", year), 
			fmt.Sprintf("%d", month), 
			fmt.Sprintf("%d", day),
			fmt.Sprintf("%d", hour),
			fmt.Sprintf("%d", minute),
			fmt.Sprintf("%d", second),
		)
		want_pattern, err := regexp.Compile(format)
		if err != nil {
			t.Fatal(err)
		}

		got_match := want_pattern.Match([]byte(got_str)) 
		if !got_match {
			t.Errorf("got misformatted date string '%s', want matching %s", got_str, want_pattern.String())
		}
	}
}