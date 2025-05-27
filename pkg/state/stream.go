package state

import (
	"encoding/binary"
	"errors"
	"io"
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