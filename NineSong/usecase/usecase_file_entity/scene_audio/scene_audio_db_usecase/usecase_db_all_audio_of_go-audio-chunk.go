package scene_audio_db_usecase

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

// Reader is a struct representing a data chunk. Its reader is shared with the
// container but convenience methods are provided.
type Reader struct {
	ID   [4]byte
	Size int
	R    io.Reader
	Pos  int
}

// Done makes sure the entire Reader was read.
func (ch *Reader) Done() {
	if !ch.IsFullyRead() {
		ch.drain()
	}
}

// Read implements the reader interface
func (ch *Reader) Read(p []byte) (n int, err error) {
	if ch == nil || ch.R == nil {
		return 0, errors.New("nil Reader/reader pointer")
	}
	n, err = ch.R.Read(p)
	ch.Pos += n
	return n, err
}

// ReadLE reads the Little Endian Reader data into the passed struct
func (ch *Reader) ReadLE(dst interface{}) error {
	return ch.readWithByteOrder(dst, binary.LittleEndian)
}

// ReadBE reads the Big Endian Reader data into the passed struct
func (ch *Reader) ReadBE(dst interface{}) error {
	return ch.readWithByteOrder(dst, binary.BigEndian)
}

// ReadByte reads and returns a single byte
func (ch *Reader) ReadByte() (byte, error) {
	if ch.IsFullyRead() {
		return 0, io.EOF
	}
	var r byte
	err := ch.ReadLE(&r)
	return r, err
}

// IsFullyRead checks if we're finished reading the Reader
func (ch *Reader) IsFullyRead() bool {
	if ch == nil || ch.R == nil {
		return true
	}
	return ch.Size <= ch.Pos
}

// Jump jumps ahead in the Reader
func (ch *Reader) Jump(bytesAhead int) error {
	var err error
	var n int64
	if bytesAhead > 0 {
		n, err = io.CopyN(ioutil.Discard, ch.R, int64(bytesAhead))
		ch.Pos += int(n)
	}
	return err
}

func (ch *Reader) readWithByteOrder(dst interface{}, byteOrder binary.ByteOrder) error {
	if ch == nil || ch.R == nil {
		return errors.New("nil Reader/reader pointer")
	}
	if ch.IsFullyRead() {
		return io.EOF
	}
	ch.Pos += binary.Size(dst)
	return binary.Read(ch.R, byteOrder, dst)
}

// You are probably looking to call Done() instead!
func (ch *Reader) drain() error {
	bytesAhead := ch.Size - ch.Pos
	if bytesAhead > 0 {
		_, err := io.CopyN(ioutil.Discard, ch.R, int64(bytesAhead))
		return err
	}
	return nil
}
