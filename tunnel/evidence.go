package tunnel

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"syscall"
	"time"

	C "github.com/metacubex/mihomo/constant"
)

type FlowEvidenceStage string

const (
	FlowEvidenceIngress         FlowEvidenceStage = "ingress"
	FlowEvidencePreHandleFailed FlowEvidenceStage = "preHandleFailed"
	FlowEvidenceRouteResolved   FlowEvidenceStage = "routeResolved"
	FlowEvidenceRouteFailed     FlowEvidenceStage = "routeFailed"
	FlowEvidenceDialStarted     FlowEvidenceStage = "dialStarted"
	FlowEvidenceDialFinished    FlowEvidenceStage = "dialFinished"
)

type FlowEvidence struct {
	Stage       FlowEvidenceStage
	At          time.Time
	Network     C.NetWork
	InboundType C.Type
	InboundName string
	SourceIP    netip.Addr
	SourcePort  uint16
	TargetIP    netip.Addr
	TargetPort  uint16
	TargetHost  string
	UID         uint32
	RuleType    string
	Outbound    string
	ErrorClass  string
	Duration    time.Duration
}

type FlowEvidenceNotify func(event FlowEvidence)

var DefaultFlowEvidenceNotify FlowEvidenceNotify

func notifyFlowEvidence(stage FlowEvidenceStage, metadata *C.Metadata, rule C.Rule, proxy C.ProxyAdapter, err error, duration time.Duration) {
	notify := DefaultFlowEvidenceNotify
	if notify == nil || metadata == nil {
		return
	}
	event := FlowEvidence{
		Stage:       stage,
		At:          time.Now(),
		Network:     metadata.NetWork,
		InboundType: metadata.Type,
		InboundName: metadata.InName,
		SourceIP:    metadata.SrcIP,
		SourcePort:  metadata.SrcPort,
		TargetIP:    metadata.DstIP,
		TargetPort:  metadata.DstPort,
		TargetHost:  metadata.Host,
		UID:         metadata.Uid,
		ErrorClass:  flowErrorClass(err),
		Duration:    duration,
	}
	if rule != nil {
		event.RuleType = rule.RuleType().String()
	}
	if proxy != nil {
		// The subscription reporter keys per-node stats on this name and matches
		// it against leaf-proxy labels, so record the dialed leaf, not the group.
		leaf := proxy
		for inner := leaf.Unwrap(metadata, false); inner != nil; inner = leaf.Unwrap(metadata, false) {
			leaf = inner
		}
		event.Outbound = leaf.Name()
	}
	notify(event)
}

func flowErrorClass(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "refused"
	case errors.Is(err, syscall.ECONNRESET):
		return "reset"
	case errors.Is(err, syscall.ENETUNREACH), errors.Is(err, syscall.EHOSTUNREACH):
		return "unreachable"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns"
	}
	return "other"
}
