package core

import (
	"time"
)

type Status struct {
	Conditions []Condition
	Phase      string
}

// SetCondition 设置状态
func (s *Status) SetCondition(condType string, condStatus string) {
	for index, condition := range s.Conditions {
		if condition.Type == condType {
			s.Conditions[index].Status = condStatus
			s.Conditions[index].LastTransitionTime = time.Now()
			return
		}
	}
	s.Conditions = append(s.Conditions, Condition{
		Type:               condType,
		Status:             condStatus,
		LastTransitionTime: time.Now(),
	})
}

// UnsetCondition 移除状态
func (s *Status) UnsetCondition(condType string) {
	for index, condition := range s.Conditions {
		if condition.Type == condType {
			s.Conditions = append(s.Conditions[:index], s.Conditions[index+1:]...)
		}
	}
}

// GetCondition 获取状态
func (s *Status) GetCondition(condType string) string {
	for _, condition := range s.Conditions {
		if condition.Type == condType {
			return condition.Status
		}
	}
	return ""
}

func NewStatus() Status {
	return Status{
		Conditions: make([]Condition, 0),
		Phase:      PhaseWaiting,
	}
}
