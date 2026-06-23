package slug_test

import (
	"testing"

	"github.com/kidkender/prohealth-api/internal/slug"
)

func TestFrom(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Bệnh Viện Đa Khoa Trung Ương", "benh-vien-da-khoa-trung-uong"},
		{"Bệnh Viện Nhi Đồng 1", "benh-vien-nhi-dong-1"},
		{"ProHealth General Hospital", "prohealth-general-hospital"},
		{"  Bệnh   Viện  Bạch Mai  ", "benh-vien-bach-mai"},
		{"Bệnh Viện Đại Học Y Dược TP.HCM", "benh-vien-dai-hoc-y-duoc-tp-hcm"},
		{"Saint Paul Hospital & Clinic", "saint-paul-hospital-clinic"},
	}

	for _, c := range cases {
		got := slug.From(c.input)
		if got != c.want {
			t.Errorf("From(%q)\n  got:  %q\n  want: %q", c.input, got, c.want)
		}
	}
}
