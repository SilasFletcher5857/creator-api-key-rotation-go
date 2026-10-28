package rotation

import "errors"

type Delivery struct {
	AssetID    string
	Subscriber string
	Processing string
}

type Plan struct {
	GraceHours int
	Deliveries []Delivery
}

func NewPlan(graceHours int, deliveries []Delivery) (Plan, error) {
	if graceHours < 1 {
		return Plan{}, errors.New("grace period must be positive")
	}
	if len(deliveries) == 0 {
		return Plan{}, errors.New("at least one delivery is required")
	}
	return Plan{GraceHours: graceHours, Deliveries: deliveries}, nil
}

func (p Plan) CutoverReady() bool {
	for _, delivery := range p.Deliveries {
		if delivery.AssetID == "" || delivery.Subscriber == "" || delivery.Processing != "complete" {
			return false
		}
	}
	return true
}
