package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	mihomoHttp "github.com/metacubex/mihomo/component/http"
	"github.com/metacubex/mihomo/log"

	"github.com/metacubex/http"
)

var (
	GeoUpdateHook func(geoType string, updating bool, skipped bool, updateErr error)
)

func sendGeoUpdateStatus(geoType string, updating bool, skipped bool, updateErr error) {
	if GeoUpdateHook != nil {
		GeoUpdateHook(geoType, updating, skipped, updateErr)
	}
}

// meta-rules-dat publishes a digest next to every database; a match means the
// multi-megabyte body would be byte-identical, so it is never fetched.
var geoDigestSuffixes = []string{".sha256sum", ".sha256"}

const (
	geoDigestTimeout = time.Second * 15
	geoDigestLimit   = 1 << 12
)

// A skipped download still refreshes mtime, which is what schedules the next run.
func skipGeoDownload(url string, path string) bool {
	local, err := os.ReadFile(path)
	if err != nil || len(local) == 0 {
		return false
	}
	remote, ok := fetchGeoDigest(url)
	if !ok || remote != hex.EncodeToString(sha256Sum(local)) {
		return false
	}
	now := time.Now()
	_ = os.Chtimes(path, now, now)
	log.Infoln("[GEO] %s matches the published digest, download skipped", path)
	return true
}

func sha256Sum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func fetchGeoDigest(url string) (string, bool) {
	for _, suffix := range geoDigestSuffixes {
		digest, ok := requestGeoDigest(url + suffix)
		if ok {
			return digest, true
		}
	}
	return "", false
}

func requestGeoDigest(url string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), geoDigestTimeout)
	defer cancel()

	resp, err := mihomoHttp.HttpRequest(ctx, url, http.MethodGet, nil, nil)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, geoDigestLimit))
	if err != nil {
		return "", false
	}
	fields := strings.Fields(string(body))
	if len(fields) == 0 {
		return "", false
	}
	digest := strings.ToLower(fields[0])
	if len(digest) != sha256.Size*2 {
		return "", false
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return "", false
	}
	return digest, true
}

var (
	geoUpdateMutex  sync.Mutex
	geoUpdateCancel context.CancelFunc
)

func StopGeoUpdater() {
	geoUpdateMutex.Lock()
	defer geoUpdateMutex.Unlock()
	stopGeoUpdater()
}

func stopGeoUpdater() {
	if geoUpdateCancel == nil {
		return
	}
	geoUpdateCancel()
	geoUpdateCancel = nil
}

func RegisterGeoUpdaterWithCancel() {
	geoUpdateMutex.Lock()
	defer geoUpdateMutex.Unlock()

	stopGeoUpdater()

	if updateInterval <= 0 {
		log.Infoln("[GEO] Invalid update interval: %d", updateInterval)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	geoUpdateCancel = cancel
	registerGeoUpdater(ctx)
}
