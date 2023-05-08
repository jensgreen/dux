package files

import "testing"

func Test_HumanizeIEC(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0B"},
		{1, "1B"},
		{999, "999B"}, // TODO want 1G I guess
		{1 * Ki, "1.0K"},
		{2*Ki + 1, "2.0K"},
		{234 * Ki, "234.0K"},
		{Gi - 1, "1024.0M"}, // TODO want 1G I guess
		{Gi, "1.0G"},
		{99*Gi + 1, "99.0G"},
		{5 * Ei, "5.0E"}, // int64 does almost overflow here...
	}
	for _, tt := range tests {
		name := tt.want
		t.Run(name, func(t *testing.T) {
			if got := HumanizeIEC(tt.n); got != tt.want {
				t.Errorf("humanize(%v) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}
