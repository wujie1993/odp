package v1

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wujie1993/waves/pkg/orm/core"
	"github.com/wujie1993/waves/pkg/orm/registry"
)

type Audit struct {
	core.BaseApiObj `json:",inline" yaml:",inline"`
	Spec            AuditSpec
}

type AuditSpec struct {
	ResourceRef ResourceRef
	Action      string
	Msg         string
	SourceIP    string
	ReqBody     string
	RespBody    string
	StatusCode  int
}

func (obj Audit) SpecEncode() ([]byte, error) {
	return json.Marshal(&obj.Spec)
}

func (obj *Audit) SpecDecode(data []byte) error {
	return json.Unmarshal(data, &obj.Spec)
}

func (obj Audit) SpecHash() string {
	data, _ := json.Marshal(&obj.Spec)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

type AuditRegistry struct {
	registry.Registry
}

func (r AuditRegistry) Record(audit *Audit) error {
	var namespaceMsg, shortNameMsg, actionMsg, descMsg string

	if audit.Spec.Msg != "" {
		descMsg = ", 备注: " + audit.Spec.Msg
	}

	if audit.Spec.ResourceRef.Namespace != "" {
		namespaceMsg = "在命名空间 " + audit.Spec.ResourceRef.Namespace + " 下"
	}

	if shortName := audit.Metadata.Annotations["ShortName"]; shortName != "" {
		shortNameMsg = "(" + shortName + ")"
	}

	switch audit.Spec.Action {
	case core.AuditActionCreate:
		actionMsg = "创建"
	case core.AuditActionUpdate:
		actionMsg = "更新"
	case core.AuditActionDelete:
		actionMsg = "删除"
	default:
		return errors.New("unsupport method")
	}

	audit.Metadata.Name = fmt.Sprintf("%d", time.Now().UnixNano())
	audit.Spec.Msg = namespaceMsg + actionMsg + core.GetKindMsg(audit.Spec.ResourceRef.Kind) + " " + audit.Spec.ResourceRef.Name + shortNameMsg + descMsg

	if _, err := r.Create(context.TODO(), audit); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func NewAudit() *Audit {
	audit := new(Audit)
	audit.Init(ApiVersion, core.KindAudit)
	return audit
}

func NewAuditRegistry() AuditRegistry {
	return AuditRegistry{
		Registry: registry.NewRegistry(newGVK(core.KindAudit), false),
	}
}
