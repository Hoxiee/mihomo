package adapter

import (
	"context"
	"time"

	"github.com/metacubex/mihomo/common/utils"
	C "github.com/metacubex/mihomo/constant"
)

// timedOut follows hub/route's getProxyDelay, which answers 504 rather than 503
// once the test's context is done.
type UrlTestCheck func(url string, name string, delay uint16, timedOut bool)

var UrlTestHook UrlTestCheck

type DialResultCheck func(name string, source string, err error, elapsed time.Duration)

var DialResultHook DialResultCheck

var DialPolicyHook func(proxy *Proxy, metadata *C.Metadata) error

func checkDialPolicy(proxy *Proxy, metadata *C.Metadata) error {
	if DialPolicyHook != nil {
		return DialPolicyHook(proxy, metadata)
	}
	return nil
}

// URLTest returns err == nil for any completed exchange and keeps the
// expected-status verdict in its per-url state, so read that back instead.
func (p *Proxy) URLTestStatus(
	ctx context.Context,
	url string,
	expectedStatus utils.IntRanges[uint16],
) (delay uint16, satisfied bool, err error) {
	delay, err = p.URLTest(ctx, url, expectedStatus)
	if err != nil {
		return delay, false, err
	}
	return delay, p.AliveForTestUrl(url), nil
}
