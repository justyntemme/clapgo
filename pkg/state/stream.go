package state

// #include <stdint.h>
// #include <string.h>
//
// typedef struct clap_istream {
//    void *ctx;
//    int64_t (*read)(const struct clap_istream *stream, void *buffer, uint64_t size);
// } clap_istream_t;
//
// typedef struct clap_ostream {
//    void *ctx;
//    int64_t (*write)(const struct clap_ostream *stream, const void *buffer, uint64_t size);
// } clap_ostream_t;
//
// static int64_t clap_stream_read(const clap_istream_t *stream, void *buffer, uint64_t size) {
//     return stream->read(stream, buffer, size);
// }
//
// static int64_t clap_stream_write(const clap_ostream_t *stream, const void *buffer, uint64_t size) {
//     return stream->write(stream, buffer, size);
// }
import "C"
import (
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

// Common stream errors
var (
	ErrStreamClosed = errors.New("stream is closed")
	ErrReadFailed   = errors.New("read failed")
	ErrWriteFailed  = errors.New("write failed")
)

// InputStream provides methods for reading binary data
type InputStream struct {
	reader io.Reader
	err    error
}

// NewInputStream creates a new input stream
func NewInputStream(reader io.Reader) *InputStream {
	return &InputStream{reader: reader}
}

// Error returns any error that occurred during reading
func (s *InputStream) Error() error {
	return s.err
}

// ReadUint32 reads a uint32 from the stream
func (s *InputStream) ReadUint32() (uint32, error) {
	if s.err != nil {
		return 0, s.err
	}
	
	var value uint32
	err := binary.Read(s.reader, binary.LittleEndian, &value)
	if err != nil {
		s.err = err
		return 0, err
	}
	
	return value, nil
}

// ReadUint64 reads a uint64 from the stream
func (s *InputStream) ReadUint64() (uint64, error) {
	if s.err != nil {
		return 0, s.err
	}
	
	var value uint64
	err := binary.Read(s.reader, binary.LittleEndian, &value)
	if err != nil {
		s.err = err
		return 0, err
	}
	
	return value, nil
}

// ReadFloat64 reads a float64 from the stream
func (s *InputStream) ReadFloat64() (float64, error) {
	if s.err != nil {
		return 0, s.err
	}
	
	var value float64
	err := binary.Read(s.reader, binary.LittleEndian, &value)
	if err != nil {
		s.err = err
		return 0, err
	}
	
	return value, nil
}

// ReadBytes reads a byte slice from the stream
func (s *InputStream) ReadBytes(n int) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	
	data := make([]byte, n)
	_, err := io.ReadFull(s.reader, data)
	if err != nil {
		s.err = err
		return nil, err
	}
	
	return data, nil
}

// ReadString reads a length-prefixed string from the stream
func (s *InputStream) ReadString() (string, error) {
	length, err := s.ReadUint32()
	if err != nil {
		return "", err
	}
	
	if length == 0 {
		return "", nil
	}
	
	data, err := s.ReadBytes(int(length))
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// Read reads raw bytes from the stream (for compatibility)
func (s *InputStream) Read(data []byte) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	
	n, err := s.reader.Read(data)
	if err != nil {
		s.err = err
		return n, err
	}
	
	return n, nil
}

// OutputStream provides methods for writing binary data
type OutputStream struct {
	writer io.Writer
	err    error
}

// NewOutputStream creates a new output stream
func NewOutputStream(writer io.Writer) *OutputStream {
	return &OutputStream{writer: writer}
}

// Error returns any error that occurred during writing
func (s *OutputStream) Error() error {
	return s.err
}

// WriteUint32 writes a uint32 to the stream
func (s *OutputStream) WriteUint32(value uint32) error {
	if s.err != nil {
		return s.err
	}
	
	err := binary.Write(s.writer, binary.LittleEndian, value)
	if err != nil {
		s.err = err
		return err
	}
	
	return nil
}

// WriteUint64 writes a uint64 to the stream
func (s *OutputStream) WriteUint64(value uint64) error {
	if s.err != nil {
		return s.err
	}
	
	err := binary.Write(s.writer, binary.LittleEndian, value)
	if err != nil {
		s.err = err
		return err
	}
	
	return nil
}

// WriteFloat64 writes a float64 to the stream
func (s *OutputStream) WriteFloat64(value float64) error {
	if s.err != nil {
		return s.err
	}
	
	err := binary.Write(s.writer, binary.LittleEndian, value)
	if err != nil {
		s.err = err
		return err
	}
	
	return nil
}

// WriteBytes writes a byte slice to the stream
func (s *OutputStream) WriteBytes(data []byte) error {
	if s.err != nil {
		return s.err
	}
	
	_, err := s.writer.Write(data)
	if err != nil {
		s.err = err
		return err
	}
	
	return nil
}

// WriteString writes a length-prefixed string to the stream
func (s *OutputStream) WriteString(value string) error {
	// Write length
	err := s.WriteUint32(uint32(len(value)))
	if err != nil {
		return err
	}
	
	// Write data
	if len(value) > 0 {
		return s.WriteBytes([]byte(value))
	}
	
	return nil
}

// Write writes raw bytes to the stream (for compatibility)
func (s *OutputStream) Write(data []byte) (int, error) {
	err := s.WriteBytes(data)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

// C Stream Adapters - these implement io.Reader/Writer for CLAP streams

// ClapReader adapts a CLAP input stream to io.Reader
type ClapReader struct {
	stream *C.clap_istream_t
}

// NewClapReader creates a new CLAP stream reader
func NewClapReader(stream unsafe.Pointer) *ClapReader {
	return &ClapReader{
		stream: (*C.clap_istream_t)(stream),
	}
}

// Read implements io.Reader
func (r *ClapReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	
	bytesRead := C.clap_stream_read(r.stream, unsafe.Pointer(&p[0]), C.uint64_t(len(p)))
	if bytesRead < 0 {
		return 0, ErrReadFailed
	}
	
	if bytesRead == 0 {
		return 0, io.EOF
	}
	
	return int(bytesRead), nil
}

// ClapWriter adapts a CLAP output stream to io.Writer
type ClapWriter struct {
	stream *C.clap_ostream_t
}

// NewClapWriter creates a new CLAP stream writer
func NewClapWriter(stream unsafe.Pointer) *ClapWriter {
	return &ClapWriter{
		stream: (*C.clap_ostream_t)(stream),
	}
}

// Write implements io.Writer
func (w *ClapWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	
	bytesWritten := C.clap_stream_write(w.stream, unsafe.Pointer(&p[0]), C.uint64_t(len(p)))
	if bytesWritten < 0 {
		return 0, ErrWriteFailed
	}
	
	return int(bytesWritten), nil
}

// Convenience functions for CLAP streams

// NewClapInputStream creates an InputStream that wraps a CLAP input stream
func NewClapInputStream(stream unsafe.Pointer) *InputStream {
	reader := NewClapReader(stream)
	return NewInputStream(reader)
}

// NewClapOutputStream creates an OutputStream that wraps a CLAP output stream
func NewClapOutputStream(stream unsafe.Pointer) *OutputStream {
	writer := NewClapWriter(stream)
	return NewOutputStream(writer)
}