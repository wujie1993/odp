package v1

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Event struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            EventSpec
}

type EventSpec struct {
	ResourceRef ResourceRef
	Action      string
	Msg         string
	JobRef      string
}

func (obj Event) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Event) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Event) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type EventRegistry struct {
	registry.Registry
}

func (r EventRegistry) Record(event *Event) error {
	var shortNameMsg string
	if shortName := event.Metadata.Annotations["ShortName"]; shortName != "" {
		shortNameMsg = "(" + shortName + ")"
	}
	appendMsg := event.Spec.Msg
	phase := event.Status.Phase

	event.Metadata.Name = fmt.Sprintf("%d", time.Now().UnixNano())
	event.Spec.Msg = core.GetKindMsg(event.Spec.ResourceRef.Kind) + " " + event.Spec.ResourceRef.Name + shortNameMsg + " "
	switch event.Status.Phase {
	case core.PhaseCompleted:
		event.Spec.Msg += core.GetActionMsg(event.Spec.Action) + "完成"
	case core.PhaseFailed:
		event.Spec.Msg += core.GetActionMsg(event.Spec.Action) + "失败"
	default:
		event.Status.Phase = core.PhaseWaiting
		event.Spec.Msg += "开始" + core.GetActionMsg(event.Spec.Action)
	}

	if appendMsg != "" {
		event.Spec.Msg += "，备注：" + appendMsg
	}

	if _, err := r.Create(context.TODO(), event); err != nil {
		log.Error(err)
		return err
	}
	if phase != core.PhaseWaiting {
		if _, err := r.UpdateStatusPhase(event.Metadata.Namespace, event.Metadata.Name, phase); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func eventPreCreate(obj core.ApiObject) error {
	event := obj.(*Event)
	event.Metadata.Finalizers = []string{core.FinalizerCleanRefJob}
	return nil
}

func NewEvent() *Event {
	event := new(Event)
	event.Init(ApiVersion, core.KindEvent)
	return event
}

func NewEventRegistry() EventRegistry {
	r := EventRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindEvent), false),
	}
	r.SetPreCreateHook(eventPreCreate)
	return r
}
