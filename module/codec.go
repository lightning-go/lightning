/**
 * Created: 2019/4/20 0020
 * @author: Jason
 */

package module

import (
	"bufio"
	"encoding/binary"
	"io"
)

type Encoder struct {
	w      io.Writer
	writer *bufio.Writer
	size   [binary.MaxVarintLen64]byte
	order  binary.ByteOrder
}

func NewEncode(w io.Writer, order binary.ByteOrder) *Encoder {
	return &Encoder{
		w:      w,
		writer: bufio.NewWriterSize(w, DefaultBufferSize),
		order:	order,
	}
}

func (encoder *Encoder) Clean() {
	encoder.writer.Reset(encoder.w)
}

func (encoder *Encoder) encodeInt(data interface{}) (err error) {
	err = binary.Write(encoder.writer, encoder.order, data)
	if err != nil {
		return err
	}
	return nil
}

func (encoder *Encoder) EncodeUInt64(data uint64) (err error) {
	return encoder.encodeInt(data)
}

func (encoder *Encoder) EncodeInt32(data int32) (err error) {
	return encoder.encodeInt(data)
}

func (encoder *Encoder) EncodeUInt32(data uint32) (err error) {
	return encoder.encodeInt(data)
}

func (encoder *Encoder) EncodeInt8(data int8) (err error) {
	return encoder.encodeInt(data)
}

func (encoder *Encoder) EncodeUInt64Tiny(data uint64) (err error) {
	n := binary.PutUvarint(encoder.size[:], data)
	_, err = encoder.writer.Write(encoder.size[:n])
	if err != nil {
		return err
	}
	return nil
}

func (encoder *Encoder) EncodeInt32Tiny(data int32) (err error) {
	return encoder.EncodeUInt64Tiny(uint64(data))
}

func (encoder *Encoder) EncodeUInt32Tiny(data uint32) (err error) {
	return encoder.EncodeUInt64Tiny(uint64(data))
}

func (encoder *Encoder) EncodeInt8Tiny(data int8) (err error) {
	return encoder.EncodeUInt64Tiny(uint64(data))
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
	order  binary.ByteOrder
}

func NewDecoder(r io.Reader, order binary.ByteOrder) *Decoder {
	return &Decoder{
		r:      r,
		reader: bufio.NewReaderSize(r, DefaultBufferSize),
		order:	order,
	}
}

func (decoder *Decoder) Clean() {
	decoder.reader.Reset(decoder.r)
}

func (decoder *Decoder) decodeInt(data interface{}) (err error) {
	err = binary.Read(decoder.reader, decoder.order, data)
	if err != nil {
		return err
	}
	return nil
}

func (decoder *Decoder) DecodeUInt64() (uint64, error) {
	var v uint64
	err := decoder.decodeInt(&v)
	return v, err
}

func (decoder *Decoder) DecodeInt32() (int32, error) {
	var v int32
	err := decoder.decodeInt(&v)
	return v, err
}

func (decoder *Decoder) DecodeUInt32() (uint32, error) {
	var v uint32
	err := decoder.decodeInt(&v)
	return v, err
}

func (decoder *Decoder) DecodeInt8() (int8, error) {
	var v int8
	err := decoder.decodeInt(&v)
	return v, err
}


func (decoder *Decoder) DecodeUInt64Tiny() (uint64, error) {
	val, err := binary.ReadUvarint(decoder.reader)
	if err != nil {
		if val == 0 {
			return 0, ErrConnClosed
		}
		return 0, err
	}
	return val, nil
}

func (decoder *Decoder) DecodeInt32Tiny() (int32, error) {
	n, err := decoder.DecodeUInt64()
	return int32(n), err
}

func (decoder *Decoder) DecodeUInt32Tiny() (uint32, error) {
	n, err := decoder.DecodeUInt64()
	return uint32(n), err
}

func (decoder *Decoder) DecodeInt8Tiny() (int8, error) {
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
	if err != nil {
		if n == 0 {
			return 0, ErrConnClosed
		}
	}
	return n, err
}
