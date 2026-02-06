package mcpbackend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/jonwraymond/tooldiscovery/index"
	"github.com/jonwraymond/toolexec/run"
	"github.com/jonwraymond/toolfoundation/model"
)

// Config describes a remote MCP backend connection.
type Config struct {
	Name       string
	URL        string
	Headers    map[string]string
	MaxRetries int
	// Transport overrides URL handling when provided (used for tests).
	Transport mcp.Transport
}

// Manager owns MCP backend connections and implements run.MCPExecutor.
type Manager struct {
	mu         sync.RWMutex
	backends   map[string]*backend
	refreshMu  sync.Mutex
	refreshing bool

	rngMu sync.Mutex
	rng   *rand.Rand
}

type backend struct {
	config      Config
	client      *mcp.Client
	session     *mcp.ClientSession
	tools       []model.Tool
	mu          sync.RWMutex
	connected   bool
	lastRefresh time.Time
}

// RefreshPolicy controls MCP backend refresh behavior.
type RefreshPolicy struct {
	Interval   time.Duration
	Jitter     time.Duration
	StaleAfter time.Duration
	OnDemand   bool
}

// Refresher implements handlers.Refresher using an MCP backend manager.
type Refresher struct {
	manager *Manager
	index   index.Index
	policy  RefreshPolicy
}

// NewRefresher creates a refresher for on-demand refreshes.
func NewRefresher(manager *Manager, idx index.Index, policy RefreshPolicy) *Refresher {
	if manager == nil || idx == nil {
		return nil
	}
	return &Refresher{manager: manager, index: idx, policy: policy}
}

// MaybeRefresh triggers a refresh if the policy indicates a stale backend.
func (r *Refresher) MaybeRefresh(ctx context.Context) error {
	if r == nil || r.manager == nil || r.index == nil {
		return nil
	}
	return r.manager.MaybeRefresh(ctx, r.index, r.policy)
}

// StartLoop starts the periodic refresh loop if enabled.
func (r *Refresher) StartLoop(ctx context.Context) {
	if r == nil || r.manager == nil || r.index == nil {
		return
	}
	r.manager.StartRefreshLoop(ctx, r.index, r.policy)
}

// NewManager validates config and returns a Manager.
func NewManager(cfgs []Config) (*Manager, error) {
	manager := &Manager{
		backends: make(map[string]*backend, len(cfgs)),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, cfg := range cfgs {
		name := strings.TrimSpace(cfg.Name)
		if name == "" {
			return nil, errors.New("mcp backend name is required")
		}
		if cfg.Transport == nil && strings.TrimSpace(cfg.URL) == "" {
			return nil, fmt.Errorf("mcp backend %q url is required", name)
		}
		if _, exists := manager.backends[name]; exists {
			return nil, fmt.Errorf("mcp backend %q already registered", name)
		}
		cfg.Name = name
		manager.backends[name] = &backend{config: cfg}
	}
	return manager, nil
}

// HasBackends reports whether any backends are configured.
func (m *Manager) HasBackends() bool {
	if m == nil {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.backends) > 0
}

// ConnectAll connects all backends and caches their tools.
func (m *Manager) ConnectAll(ctx context.Context) error {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	backends := make(map[string]*backend, len(m.backends))
	for name, b := range m.backends {
		backends[name] = b
	}
	m.mu.RUnlock()

	for name, b := range backends {
		tools, err := b.fetchTools(ctx)
		if err != nil {
			return fmt.Errorf("connect backend %s: %w", name, err)
		}
		b.storeTools(tools)
	}
	return nil
}

// RegisterTools registers backend tools into the provided index.
func (m *Manager) RegisterTools(idx index.Index) error {
	if m == nil || idx == nil {
		return nil
	}
	m.mu.RLock()
	backends := make(map[string]*backend, len(m.backends))
	for name, b := range m.backends {
		backends[name] = b
	}
	m.mu.RUnlock()

	for name, b := range backends {
		tools := b.toolsSnapshot()
		if len(tools) == 0 {
			continue
		}
		if err := idx.RegisterToolsFromMCP(name, tools); err != nil {
			return fmt.Errorf("register backend %s tools: %w", name, err)
		}
	}
	return nil
}

// StartRefreshLoop periodically refreshes MCP backends and updates the index.
func (m *Manager) StartRefreshLoop(ctx context.Context, idx index.Index, policy RefreshPolicy) {
	if m == nil || idx == nil {
		return
	}
	if policy.Interval <= 0 {
		return
	}

	go func() {
		timer := time.NewTimer(m.jittered(policy.Interval, policy.Jitter))
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := m.RefreshAll(ctx, idx); err != nil {
					slog.Default().Warn("mcp backend refresh failed", "err", err)
				} else {
					slog.Default().Info("mcp backend refresh completed")
				}
				timer.Reset(m.jittered(policy.Interval, policy.Jitter))
			}
		}
	}()
}

// MaybeRefresh triggers a refresh if any backend is stale.
func (m *Manager) MaybeRefresh(ctx context.Context, idx index.Index, policy RefreshPolicy) error {
	if m == nil || idx == nil {
		return nil
	}
	if !policy.OnDemand || policy.StaleAfter <= 0 {
		return nil
	}
	if !m.anyStale(policy.StaleAfter) {
		return nil
	}
	return m.RefreshAll(ctx, idx)
}

// RefreshAll refreshes tools for all MCP backends and updates the index.
func (m *Manager) RefreshAll(ctx context.Context, idx index.Index) error {
	if m == nil || idx == nil {
		return nil
	}

	if !m.beginRefresh() {
		return nil
	}
	defer m.endRefresh()

	m.mu.RLock()
	backends := make(map[string]*backend, len(m.backends))
	for name, b := range m.backends {
		backends[name] = b
	}
	m.mu.RUnlock()

	var errs []error
	for name, b := range backends {
		oldTools := b.toolsSnapshot()
		newTools, err := b.fetchTools(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("refresh backend %s: %w", name, err))
			continue
		}
		if err := syncBackendTools(idx, name, oldTools, newTools); err != nil {
			errs = append(errs, fmt.Errorf("sync backend %s tools: %w", name, err))
			continue
		}
		b.storeTools(newTools)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (m *Manager) beginRefresh() bool {
	m.refreshMu.Lock()
	defer m.refreshMu.Unlock()
	if m.refreshing {
		return false
	}
	m.refreshing = true
	return true
}

func (m *Manager) endRefresh() {
	m.refreshMu.Lock()
	m.refreshing = false
	m.refreshMu.Unlock()
}

func (m *Manager) anyStale(staleAfter time.Duration) bool {
	now := time.Now()
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, b := range m.backends {
		if b.isStale(now, staleAfter) {
			return true
		}
	}
	return false
}

func (m *Manager) jittered(base, jitter time.Duration) time.Duration {
	if base <= 0 {
		return base
	}
	if jitter <= 0 {
		return base
	}

	maxJitter := int64(jitter)
	if maxJitter <= 0 {
		return base
	}
	max := maxJitter * 2

	var n int64
	m.rngMu.Lock()
	if m.rng == nil {
		m.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	n = m.rng.Int63n(max + 1)
	m.rngMu.Unlock()

	delta := n - maxJitter
	next := base + time.Duration(delta)
	if next <= 0 {
		return base
	}
	return next
}

// ToolsSnapshot returns a copy of cached tools per backend.
func (m *Manager) ToolsSnapshot() map[string][]model.Tool {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string][]model.Tool, len(m.backends))
	for name, b := range m.backends {
		out[name] = b.toolsSnapshot()
	}
	return out
}

// CallTool executes a tool on a remote MCP backend.
func (m *Manager) CallTool(ctx context.Context, serverName string, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	backend, err := m.lookupBackend(serverName)
	if err != nil {
		return nil, err
	}
	if err := backend.ensureConnected(ctx); err != nil {
		return nil, err
	}

	backend.mu.RLock()
	session := backend.session
	backend.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("mcp backend %q not connected", serverName)
	}
	return session.CallTool(ctx, params)
}

// CallToolStream is not supported by the MCP SDK client yet.
func (m *Manager) CallToolStream(context.Context, string, *mcp.CallToolParams) (<-chan run.StreamEvent, error) {
	return nil, run.ErrStreamNotSupported
}

// Close disconnects all backends.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	backends := make(map[string]*backend, len(m.backends))
	for name, b := range m.backends {
		backends[name] = b
	}
	m.mu.RUnlock()

	for name, b := range backends {
		if err := b.disconnect(); err != nil {
			return fmt.Errorf("disconnect backend %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) lookupBackend(name string) (*backend, error) {
	if m == nil {
		return nil, errors.New("mcp backend manager not configured")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("mcp backend name is required")
	}
	m.mu.RLock()
	backend := m.backends[name]
	m.mu.RUnlock()
	if backend == nil {
		return nil, fmt.Errorf("mcp backend %q not found", name)
	}
	return backend, nil
}

func (b *backend) ensureConnected(ctx context.Context) error {
	b.mu.RLock()
	connected := b.connected
	b.mu.RUnlock()
	if connected {
		return nil
	}
	return b.connect(ctx)
}

func (b *backend) connect(ctx context.Context) error {
	b.mu.Lock()
	if b.connected {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	transport, err := b.transport()
	if err != nil {
		return err
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "metatools-mcp-backend"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.client = client
	b.session = session
	b.connected = true
	b.lastRefresh = time.Now()
	b.mu.Unlock()
	return nil
}

func (b *backend) fetchTools(ctx context.Context) ([]model.Tool, error) {
	if err := b.ensureConnected(ctx); err != nil {
		return nil, err
	}

	b.mu.RLock()
	session := b.session
	serverName := b.config.Name
	b.mu.RUnlock()
	if session == nil {
		return nil, fmt.Errorf("mcp backend %q not connected", serverName)
	}

	res, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	tools := make([]model.Tool, 0, len(res.Tools))
	for _, tool := range res.Tools {
		if tool == nil {
			continue
		}
		normalized := normalizeMCPTool(serverName, model.Tool{Tool: *tool})
		tools = append(tools, normalized)
	}
	return tools, nil
}

func (b *backend) disconnect() error {
	b.mu.Lock()
	if !b.connected {
		b.mu.Unlock()
		return nil
	}
	session := b.session
	b.client = nil
	b.session = nil
	b.connected = false
	b.mu.Unlock()

	if session != nil {
		return session.Close()
	}
	return nil
}

func (b *backend) toolsSnapshot() []model.Tool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.tools) == 0 {
		return nil
	}
	out := make([]model.Tool, len(b.tools))
	copy(out, b.tools)
	return out
}

func (b *backend) storeTools(tools []model.Tool) {
	b.mu.Lock()
	b.tools = tools
	b.lastRefresh = time.Now()
	b.mu.Unlock()
}

func (b *backend) isStale(now time.Time, staleAfter time.Duration) bool {
	b.mu.RLock()
	last := b.lastRefresh
	connected := b.connected
	b.mu.RUnlock()
	if !connected {
		return true
	}
	if last.IsZero() {
		return true
	}
	return now.Sub(last) > staleAfter
}

func (b *backend) transport() (mcp.Transport, error) {
	if b.config.Transport != nil {
		return b.config.Transport, nil
	}
	if strings.TrimSpace(b.config.URL) == "" {
		return nil, errors.New("backend URL is required")
	}

	parsed, err := url.Parse(b.config.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid backend URL: %w", err)
	}

	httpClient := httpClientWithHeaders(b.config.Headers)

	switch parsed.Scheme {
	case "http", "https":
		return &mcp.StreamableClientTransport{
			Endpoint:   b.config.URL,
			HTTPClient: httpClient,
			MaxRetries: b.config.MaxRetries,
		}, nil
	case "sse":
		parsed.Scheme = "http"
		return &mcp.SSEClientTransport{
			Endpoint:   parsed.String(),
			HTTPClient: httpClient,
		}, nil
	case "stdio":
		return &mcp.StdioTransport{}, nil
	default:
		return nil, fmt.Errorf("unsupported backend URL scheme %q", parsed.Scheme)
	}
}

func httpClientWithHeaders(headers map[string]string) *http.Client {
	if len(headers) == 0 {
		return nil
	}
	clone := make(map[string]string, len(headers))
	for k, v := range headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		clone[k] = v
	}
	if len(clone) == 0 {
		return nil
	}
	return &http.Client{
		Transport: &headerRoundTripper{
			base:    http.DefaultTransport,
			headers: clone,
		},
	}
}

type headerRoundTripper struct {
	base    http.RoundTripper
	headers map[string]string
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	for k, v := range h.headers {
		if clone.Header.Get(k) == "" {
			clone.Header.Set(k, v)
		}
	}
	return h.base.RoundTrip(clone)
}

func normalizeMCPTool(serverName string, tool model.Tool) model.Tool {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		return tool
	}
	prefix := "mcp." + serverName
	namespace := strings.TrimSpace(tool.Namespace)
	switch {
	case namespace == "":
		tool.Namespace = prefix
	case strings.HasPrefix(namespace, prefix):
		tool.Namespace = namespace
	default:
		tool.Namespace = prefix + "." + namespace
	}

	tags := append([]string{}, tool.Tags...)
	tags = append(tags, "backend.mcp", "server."+serverName)
	tool.Tags = model.NormalizeTags(tags)
	return tool
}

func syncBackendTools(idx index.Index, serverName string, oldTools []model.Tool, newTools []model.Tool) error {
	backend := model.ToolBackend{
		Kind: model.BackendKindMCP,
		MCP:  &model.MCPBackend{ServerName: serverName},
	}

	newIDs := make(map[string]struct{}, len(newTools))
	for _, tool := range newTools {
		newIDs[tool.ToolID()] = struct{}{}
		if err := idx.RegisterTool(tool, backend); err != nil {
			return err
		}
	}

	for _, tool := range oldTools {
		if _, ok := newIDs[tool.ToolID()]; ok {
			continue
		}
		_ = idx.UnregisterBackend(tool.ToolID(), model.BackendKindMCP, serverName)
	}
	return nil
}
