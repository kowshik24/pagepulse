package app

import "testing"

func TestParseConfigRejectsTooSmallSampleInterval(t *testing.T) {
	_, err := ParseConfig([]string{"--sample-interval", "100ms"})
	if err == nil {
		t.Fatal("expected error for too-small sample interval")
	}
}
