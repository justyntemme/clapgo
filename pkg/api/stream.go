package api

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
// static int64_t stream_read(const clap_istream_t *stream, void *buffer, uint64_t size) {
//     return stream->read(stream, buffer, size);
// }
//
// static int64_t stream_write(const clap_ostream_t *stream, const void *buffer, uint64_t size) {
//     return stream->write(stream, buffer, size);
// }
import "C"
import (
	"encoding/binary"
	"errors"
	"unsafe"
)

// Stream errors
var (
	ErrStreamRead  = errors.New("stream read error")
	ErrStreamWrite = errors.New("stream write error")
)

// InputStream wraps a CLAP input stream for reading data
type InputStream struct {
	stream *C.clap_istream_t
}

// NewInputStream creates a new input stream wrapper
func NewInputStream(stream unsafe.Pointer) *InputStream {
	return &InputStream{
		stream: (*C.clap_istream_t)(stream),
	}
}

// Read reads data from the stream
func (s *InputStream) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	
	n := C.stream_read(s.stream, unsafe.Pointer(&data[0]), C.uint64_t(len(data)))
	if n < 0 {
		return 0, ErrStreamRead
	}
	return int(n), nil
}

// ReadUint32 reads a uint32 from the stream
func (s *InputStream) ReadUint32() (uint32, error) {
	var buf [4]byte
	n, err := s.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, ErrStreamRead
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

// ReadFloat64 reads a float64 from the stream
func (s *InputStream) ReadFloat64() (float64, error) {
	var buf [8]byte
	n, err := s.Read(buf[:])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, ErrStreamRead
	}
	bits := binary.LittleEndian.Uint64(buf[:])
	return *(*float64)(unsafe.Pointer(&bits)), nil
}

// ReadString reads a string from the stream (prefixed with uint32 length)
func (s *InputStream) ReadString() (string, error) {
	length, err := s.ReadUint32()
	if err != nil {
		return "", err
	}
	if length == 0 {
		return "", nil
	}
	
	buf := make([]byte, length)
	n, err := s.Read(buf)
	if err != nil {
		return "", err
	}
	if uint32(n) != length {
		return "", ErrStreamRead
	}
	return string(buf), nil
}

// OutputStream wraps a CLAP output stream for writing data
type OutputStream struct {
	stream *C.clap_ostream_t
}

// NewOutputStream creates a new output stream wrapper
func NewOutputStream(stream unsafe.Pointer) *OutputStream {
	return &OutputStream{
		stream: (*C.clap_ostream_t)(stream),
	}
}

// Write writes data to the stream
func (s *OutputStream) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	
	n := C.stream_write(s.stream, unsafe.Pointer(&data[0]), C.uint64_t(len(data)))
	if n < 0 {
		return 0, ErrStreamWrite
	}
	return int(n), nil
}

// WriteUint32 writes a uint32 to the stream
func (s *OutputStream) WriteUint32(v uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	n, err := s.Write(buf[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return ErrStreamWrite
	}
	return nil
}

// WriteFloat64 writes a float64 to the stream
func (s *OutputStream) WriteFloat64(v float64) error {
	var buf [8]byte
	bits := *(*uint64)(unsafe.Pointer(&v))
	binary.LittleEndian.PutUint64(buf[:], bits)
	n, err := s.Write(buf[:])
	if err != nil {
		return err
	}
	if n != 8 {
		return ErrStreamWrite
	}
	return nil
}

// WriteString writes a string to the stream (prefixed with uint32 length)
func (s *OutputStream) WriteString(v string) error {
	if err := s.WriteUint32(uint32(len(v))); err != nil {
		return err
	}
	if len(v) == 0 {
		return nil
	}
	
	n, err := s.Write([]byte(v))
	if err != nil {
		return err
	}
	if n != len(v) {
		return ErrStreamWrite
	}
	return nil
}