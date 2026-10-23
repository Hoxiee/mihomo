package tunnel

import (
	"context"
	"errors"
	"net/netip"
	"syscall"
	"testing"
	"time"

	C "github.com/metacubex/mihomo/constant"
)

func TestNotifyFlowEvidenceCopiesMetadata(t *testing.T) {
	previous := DefaultFlowEvidenceNotify
	t.Cleanup(func() { DefaultFlowEvidenceNotify = previous })

	var got FlowEvidence
	DefaultFlowEvidenceNotify = func(event FlowEvidence) {
		got = event
	}
	metadata := &C.Metadata{
		NetWork: C.TCP,
		Type:    C.TUN,
		SrcIP:   netip.MustParseAddr("10.0.0.2"),
		SrcPort: 42000,
		DstIP:   netip.MustParseAddr("203.0.113.8"),
		DstPort: 443,
		Host:    "example.test",
		InName:  "tun0",
		Uid:     10042,
	}

	notifyFlowEvidence(FlowEvidenceIngress, metadata, nil, nil, nil, time.Second)
	metadata.Host = "changed.test"
	metadata.SrcPort = 1

	if got.Stage != FlowEvidenceIngress || got.SourcePort != 42000 || got.TargetHost != "example.test" {
		t.Fatalf("unexpected copied event: %+v", got)
	}
	if got.Network != C.TCP || got.InboundType != C.TUN || got.Duration != time.Second {
		t.Fatalf("missing normalized metadata: %+v", got)
	}
}

func TestFlowErrorClass(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "none", want: ""},
		{name: "cancelled", err: context.Canceled, want: "cancelled"},
		{name: "deadline", err: context.DeadlineExceeded, want: "timeout"},
		{name: "refused", err: syscall.ECONNREFUSED, want: "refused"},
		{name: "reset", err: syscall.ECONNRESET, want: "reset"},
		{name: "unreachable", err: syscall.ENETUNREACH, want: "unreachable"},
		{name: "wrapped", err: errors.Join(errors.New("dial"), syscall.EHOSTUNREACH), want: "unreachable"},
		{name: "other", err: errors.New("opaque"), want: "other"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := flowErrorClass(test.err); got != test.want {
				t.Fatalf("flowErrorClass() = %q, want %q", got, test.want)
			}
		})
	}
}
