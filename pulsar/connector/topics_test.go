package connector

import "testing"

func TestGetTopicName(t *testing.T) {
	topic := "persistent://public/default/attack-chain-scan-state-v1"
	want := "attack-chain-scan-state-v1"
	got := GetTopicName(topic)
	if got != want {
		t.Errorf("GetTopicName() = %v, want %v", got, want)
	}

	topic = ""
	want = ""
	got = GetTopicName(topic)
	if got != want {
		t.Errorf("GetTopicName() = %v, want %v", got, want)
	}
}
