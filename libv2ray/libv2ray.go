package libv2ray

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/xtls/xray-core/app/stats"
	corefilesystem "github.com/xtls/xray-core/common/platform/filesystem"
	"github.com/xtls/xray-core/core"
	corestats "github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/infra/conf/serial"

	// Register all protocols and features for Xray-core
	_ "github.com/xtls/xray-core/main/distro/all"
)

// InitCoreEnv initializes asset paths and runtime environment for Android
func InitCoreEnv(envPath string, key string) {
	if envPath != "" {
		_ = os.Setenv("xray.location.asset", envPath)
		_ = os.Setenv("v2ray.location.asset", envPath)
		corefilesystem.NewFileReader = func(path string) (io.ReadCloser, error) {
			targetPath := filepath.Join(envPath, path)
			return os.Open(targetPath)
		}
	}
	if key != "" {
		_ = os.Setenv("xray.xudp.basekey", key)
	}
}

// CoreCallbackHandler handles lifecycle events emitted by Xray core
type CoreCallbackHandler interface {
	OnEmitStatus(status int64, message string) int64
	Shutdown() int64
	Startup() int64
}

// CoreController manages the active Xray-core instance
type CoreController struct {
	mu        sync.Mutex
	instance  *core.Instance
	callback  CoreCallbackHandler
	isRunning bool
}

// NewCoreController creates a new CoreController instance
func NewCoreController(s CoreCallbackHandler) *CoreController {
	return &CoreController{
		callback: s,
	}
}

// StartLoop initializes and starts the Xray core loop with the provided JSON config
func (c *CoreController) StartLoop(configContent string, fd int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return errors.New("core is already running")
	}

	if fd > 0 {
		_ = os.Setenv("XRAY_TUN_FD", strconv.Itoa(fd))
	}

	cfg, err := serial.DecodeJSONConfig(strings.NewReader(configContent))
	if err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	server, err := core.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create core instance: %w", err)
	}

	if err := server.Start(); err != nil {
		_ = server.Close()
		return fmt.Errorf("failed to start core instance: %w", err)
	}

	c.instance = server
	c.isRunning = true

	if c.callback != nil {
		c.callback.Startup()
	}

	return nil
}

// StopLoop gracefully terminates the Xray core instance
func (c *CoreController) StopLoop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning || c.instance == nil {
		return nil
	}

	err := c.instance.Close()
	c.instance = nil
	c.isRunning = false

	if c.callback != nil {
		c.callback.Shutdown()
	}

	return err
}

// IsRunning returns whether the core loop is currently active
func (c *CoreController) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isRunning
}

// QueryAllOutboundTrafficStats queries traffic statistics from Xray stats manager.
// Returns formatted string: "tag,direction,value;tag,direction,value;"
func (c *CoreController) QueryAllOutboundTrafficStats() string {
	c.mu.Lock()
	server := c.instance
	running := c.isRunning
	c.mu.Unlock()

	if !running || server == nil {
		return ""
	}

	statsFeature := server.GetFeature(corestats.ManagerType())
	if statsFeature == nil {
		return ""
	}
	manager, ok := statsFeature.(corestats.Manager)
	if !ok {
		return ""
	}

	var sb strings.Builder
	// Standard Xray counter pattern: outbound>>><tag>>>>traffic>>><downlink|uplink>
	for _, direction := range []string{"downlink", "uplink"} {
		counterName := fmt.Sprintf("outbound>>>proxy>>>traffic>>>%s", direction)
		counter := manager.GetCounter(counterName)
		if counter != nil {
			val := counter.Set(0) // Read and delta-reset
			sb.WriteString(fmt.Sprintf("proxy,%s,%d;", direction, val))
		}
	}

	return sb.String()
}

// MeasureOutboundDelay tests latency through the configured outbound in milliseconds
func MeasureOutboundDelay(configContent string, testURL string) (int64, error) {
	if testURL == "" {
		testURL = "https://cp.cloudflare.com/generate_204"
	}

	cfg, err := serial.DecodeJSONConfig(strings.NewReader(configContent))
	if err != nil {
		return -1, fmt.Errorf("decode config error: %w", err)
	}

	server, err := core.New(cfg)
	if err != nil {
		return -1, fmt.Errorf("new core error: %w", err)
	}
	defer server.Close()

	if err := server.Start(); err != nil {
		return -1, fmt.Errorf("start core error: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return -1, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()
	return duration, nil
}
