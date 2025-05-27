package state

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"
)

// ErrContextCanceled is returned when an operation is canceled via context
var ErrContextCanceled = errors.New("operation canceled")

// SaveWithContext saves state with context support for cancellation
func (m *Manager) SaveWithContext(ctx context.Context, w io.Writer, state *State) error {
	// Check context before starting
	select {
	case <-ctx.Done():
		return ErrContextCanceled
	default:
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	
	// Write with periodic context checks
	const chunkSize = 4096
	for i := 0; i < len(data); i += chunkSize {
		// Check context
		select {
		case <-ctx.Done():
			return ErrContextCanceled
		default:
		}
		
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		if _, err := w.Write(data[i:end]); err != nil {
			return err
		}
	}
	
	return nil
}

// LoadWithContext loads state with context support for cancellation
func (m *Manager) LoadWithContext(ctx context.Context, r io.Reader) (*State, error) {
	// Check context before starting
	select {
	case <-ctx.Done():
		return nil, ErrContextCanceled
	default:
	}
	
	// Read with periodic context checks
	data := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	
	for {
		// Check context
		select {
		case <-ctx.Done():
			return nil, ErrContextCanceled
		default:
		}
		
		n, err := r.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	
	// Parse JSON
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	
	// Validate
	if err := m.ValidateState(&state); err != nil {
		return nil, err
	}
	
	return &state, nil
}

// SaveAsyncResult represents the result of an async save operation
type SaveAsyncResult struct {
	Error error
	Done  chan struct{}
}

// SaveAsync saves state asynchronously
func (m *Manager) SaveAsync(w io.Writer, state *State) *SaveAsyncResult {
	result := &SaveAsyncResult{
		Done: make(chan struct{}),
	}
	
	go func() {
		defer close(result.Done)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		result.Error = m.SaveWithContext(ctx, w, state)
	}()
	
	return result
}

// LoadAsyncResult represents the result of an async load operation
type LoadAsyncResult struct {
	State *State
	Error error
	Done  chan struct{}
}

// LoadAsync loads state asynchronously
func (m *Manager) LoadAsync(r io.Reader) *LoadAsyncResult {
	result := &LoadAsyncResult{
		Done: make(chan struct{}),
	}
	
	go func() {
		defer close(result.Done)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		result.State, result.Error = m.LoadWithContext(ctx, r)
	}()
	
	return result
}

// StreamWriter provides a context-aware state writer
type StreamWriter struct {
	ctx    context.Context
	writer io.Writer
	err    error
}

// NewStreamWriter creates a new context-aware stream writer
func NewStreamWriter(ctx context.Context, w io.Writer) *StreamWriter {
	return &StreamWriter{
		ctx:    ctx,
		writer: w,
	}
}

// WriteParameter writes a parameter with context checking
func (s *StreamWriter) WriteParameter(p Parameter) error {
	if s.err != nil {
		return s.err
	}
	
	select {
	case <-s.ctx.Done():
		s.err = ErrContextCanceled
		return s.err
	default:
	}
	
	data, err := json.Marshal(p)
	if err != nil {
		s.err = err
		return err
	}
	
	if _, err := s.writer.Write(data); err != nil {
		s.err = err
		return err
	}
	
	return nil
}

// Error returns any error that occurred during writing
func (s *StreamWriter) Error() error {
	return s.err
}