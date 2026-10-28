package rotation

import "testing"

func TestCutoverReadyRequiresProcessedSubscriberDeliveries(t *testing.T) {
	tests := []struct {
		name       string
		deliveries []Delivery
		want       bool
	}{
		{name: "processed asset with subscriber", deliveries: []Delivery{{AssetID: "course-17", Subscriber: "sub-9", Processing: "complete"}}, want: true},
		{name: "processing still running", deliveries: []Delivery{{AssetID: "course-17", Subscriber: "sub-9", Processing: "queued"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := NewPlan(24, tt.deliveries)
			if err != nil {
				t.Fatal(err)
			}
			if got := plan.CutoverReady(); got != tt.want {
				t.Fatalf("CutoverReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
