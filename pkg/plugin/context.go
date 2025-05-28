package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ContextKey is used for context values
type ContextKey string

const (
	// ContextKeyTimeout sets operation timeout
	ContextKeyTimeout ContextKey = "timeout"
	// ContextKeyProgress sets progress channel
	ContextKeyProgress ContextKey = "progress"
	// ContextKeyPluginID sets plugin ID for logging
	ContextKeyPluginID ContextKey = "plugin_id"
)

// WithTimeout adds a timeout to the context
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithProgress adds a progress channel to the context
func WithProgress(parent context.Context, progress chan<- ProgressUpdate) context.Context {
	return context.WithValue(parent, ContextKeyProgress, progress)
}

// WithPluginID adds plugin ID to the context for logging
func WithPluginID(parent context.Context, pluginID string) context.Context {
	return context.WithValue(parent, ContextKeyPluginID, pluginID)
}

// GetProgress retrieves the progress channel from context
func GetProgress(ctx context.Context) (chan<- ProgressUpdate, bool) {
	progress, ok := ctx.Value(ContextKeyProgress).(chan<- ProgressUpdate)
	return progress, ok
}

// GetPluginID retrieves the plugin ID from context
func GetPluginID(ctx context.Context) (string, bool) {
	pluginID, ok := ctx.Value(ContextKeyPluginID).(string)
	return pluginID, ok
}

// ReportProgress safely reports progress if a channel is available
func ReportProgress(ctx context.Context, update ProgressUpdate) {
	if progress, ok := GetProgress(ctx); ok {
		select {
		case progress <- update:
		case <-ctx.Done():
			return
		default:
			// Don't block if channel is full
		}
	}
}

// ParameterTransactionImpl implements ParameterTransaction
type ParameterTransactionImpl struct {
	ctx      context.Context
	cancel   context.CancelFunc
	changes  map[uint32]float64
	original map[uint32]float64
	plugin   *PluginBase
	mu       sync.RWMutex
	applied  bool
}

// NewParameterTransaction creates a new parameter transaction
func NewParameterTransaction(ctx context.Context, plugin *PluginBase) *ParameterTransactionImpl {
	txCtx, cancel := context.WithCancel(ctx)
	return &ParameterTransactionImpl{
		ctx:      txCtx,
		cancel:   cancel,
		changes:  make(map[uint32]float64),
		original: make(map[uint32]float64),
		plugin:   plugin,
		applied:  false,
	}
}

// SetParameter sets a parameter in the transaction
func (t *ParameterTransactionImpl) SetParameter(id uint32, value float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	select {
	case <-t.ctx.Done():
		return t.ctx.Err()
	default:
	}
	
	if t.applied {
		return fmt.Errorf("transaction already applied")
	}
	
	// Store original value if this is the first change to this parameter
	if _, exists := t.changes[id]; !exists {
		if currentValue, err := t.plugin.ParamManager.GetValue(id); err == nil {
			t.original[id] = currentValue
		}
	}
	
	t.changes[id] = value
	return nil
}

// Commit applies all parameter changes
func (t *ParameterTransactionImpl) Commit() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	select {
	case <-t.ctx.Done():
		return t.ctx.Err()
	default:
	}
	
	if t.applied {
		return fmt.Errorf("transaction already applied")
	}
	
	// Apply all changes
	for id, value := range t.changes {
		if err := t.plugin.ParamManager.SetValue(id, value); err != nil {
			// Rollback already applied changes
			t.rollbackPartial(id)
			return fmt.Errorf("failed to apply parameter %d: %w", id, err)
		}
	}
	
	t.applied = true
	t.cancel()
	return nil
}

// Rollback cancels all parameter changes
func (t *ParameterTransactionImpl) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.applied {
		// Restore original values
		for id, originalValue := range t.original {
			t.plugin.ParamManager.SetValue(id, originalValue)
		}
	}
	
	t.cancel()
	return nil
}

// rollbackPartial rolls back changes up to (but not including) the failed parameter
func (t *ParameterTransactionImpl) rollbackPartial(failedID uint32) {
	for id, originalValue := range t.original {
		if id == failedID {
			break
		}
		t.plugin.ParamManager.SetValue(id, originalValue)
	}
}

// Context returns the transaction context
func (t *ParameterTransactionImpl) Context() context.Context {
	return t.ctx
}

// ContextualStateWriter wraps a StateWriter with context support
type ContextualStateWriter struct {
	writer StateWriter
}

// NewContextualStateWriter creates a new contextual state writer
func NewContextualStateWriter(writer StateWriter) *ContextualStateWriter {
	return &ContextualStateWriter{writer: writer}
}

// Implement StateWriter interface
func (w *ContextualStateWriter) WriteUint32(v uint32) error {
	return w.writer.WriteUint32(v)
}

func (w *ContextualStateWriter) WriteFloat64(v float64) error {
	return w.writer.WriteFloat64(v)
}

func (w *ContextualStateWriter) WriteString(s string) error {
	return w.writer.WriteString(s)
}

func (w *ContextualStateWriter) WriteBytes(b []byte) error {
	return w.writer.WriteBytes(b)
}

// Implement StateWriterV2 interface
func (w *ContextualStateWriter) WriteUint32WithContext(ctx context.Context, v uint32) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return w.writer.WriteUint32(v)
	}
}

func (w *ContextualStateWriter) WriteFloat64WithContext(ctx context.Context, v float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return w.writer.WriteFloat64(v)
	}
}

func (w *ContextualStateWriter) WriteStringWithContext(ctx context.Context, s string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return w.writer.WriteString(s)
	}
}

func (w *ContextualStateWriter) WriteBytesWithContext(ctx context.Context, b []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return w.writer.WriteBytes(b)
	}
}

// ContextualStateReader wraps a StateReader with context support
type ContextualStateReader struct {
	reader StateReader
}

// NewContextualStateReader creates a new contextual state reader
func NewContextualStateReader(reader StateReader) *ContextualStateReader {
	return &ContextualStateReader{reader: reader}
}

// Implement StateReader interface
func (r *ContextualStateReader) ReadUint32() (uint32, error) {
	return r.reader.ReadUint32()
}

func (r *ContextualStateReader) ReadFloat64() (float64, error) {
	return r.reader.ReadFloat64()
}

func (r *ContextualStateReader) ReadString() (string, error) {
	return r.reader.ReadString()
}

func (r *ContextualStateReader) ReadBytes(n int) ([]byte, error) {
	return r.reader.ReadBytes(n)
}

// Implement StateReaderV2 interface
func (r *ContextualStateReader) ReadUint32WithContext(ctx context.Context) (uint32, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return r.reader.ReadUint32()
	}
}

func (r *ContextualStateReader) ReadFloat64WithContext(ctx context.Context) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return r.reader.ReadFloat64()
	}
}

func (r *ContextualStateReader) ReadStringWithContext(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return r.reader.ReadString()
	}
}

func (r *ContextualStateReader) ReadBytesWithContext(ctx context.Context, n int) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return r.reader.ReadBytes(n)
	}
}