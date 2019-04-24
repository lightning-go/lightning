/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package module

import (
	"encoding/binary"
	"bufio"
	"io"
)

type Encoder struct {
	w      io.Writer
	writer *bufio.Writer
	size   [binary.MaxVarintLen64]byte
}

func NewEocode(w io.Writer) *Encoder {
	return &Encoder{
		w:      w,
		writer: bufio.NewWriterSize(w, DefaultBufferSize),
	}
}

func (encoder *Encoder) Clean() {
	encoder.writer.Reset(encoder.w)
}

func (encoder *Encoder) EncodeUInt64(data uint64) (err error) {
	n := binary.PutUvarint(encoder.size[:], data)
	_, err = encoder.writer.Write(encoder.size[:n])
	if err != nil {
		return err
	}
	return nil
}

func (encoder *Encoder) EncodeInt32(data int32) (err error) {
	return encoder.EncodeUInt64(uint64(data))
}

func (encoder *Encoder) EncodeUInt32(data uint32) (err error) {
	return encoder.EncodeUInt64(uint64(data))
}

func (encoder *Encoder) EncodeInt8(data int8) (err error) {
	return encoder.EncodeUInt64(uint64(data))
}

func (encoder *Encoder) EncodeData(data []byte) (err error) {
	_, err = encoder.writer.Write(data)
	return err
}

func (encoder *Encoder) Flush() error {
	return encoder.writer.Flush()
}

//////////////////////////////////////////////////////////////////////

const (
	DefaultBufferSize = 8 * 1024
)

type Decoder struct {
	r      io.Reader
	reader *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      r,
		reader: bufio.NewReaderSize(r, DefaultBufferSize),
	}
}

func (decoder *Decoder) Clean() {
	decoder.reader.Reset(decoder.r)
}

func (decoder *Decoder) DecodeUInt64() (uint64, error) {
	val, err := binary.ReadUvarint(decoder.reader)
	if err != nil {
		if val == 0 {
			return 0, ErrConnClosed
		}
		return 0, err
	}
	return val, nil
}

func (decoder *Decoder) DecodeInt32() (int32, error) {
	n, err := decoder.DecodeUInt64()
	return int32(n), err
}

func (decoder *Decoder) DecodeUInt32() (uint32, error) {
	n, err := decoder.DecodeUInt64()
	return uint32(n), err
}

func (decoder *Decoder) DecodeInt8() (int8, error) {
	n, err := decoder.DecodeUInt64()
	return int8(n), err
}

func (decoder *Decoder) DecodeData(buf []byte) (n int, err error) {
	if buf == nil {
		return 0, ErrReadBuffNil
	}
	n, err = decoder.reader.Read(buf)
	if err != nil {
		if n == 0 {
			return 0, ErrConnClosed
		}
		return 0, err
	}
	return n, err
}

func (decoder *Decoder) DecodeDataFull(buf []byte) (n int, err error) {
	if buf == nil {
		return 0, ErrReadBuffNil
	}
	n, err = io.ReadFull(decoder.reader, buf)
	if n == 0 {
		return 0, ErrConnClosed
	}
	return n, err
}
