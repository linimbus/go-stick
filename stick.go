package stick

import (
	"encoding/binary"
	"fmt"
	"io"
)

const STICK_MAX_BUFF_LEN = 512*1024
const STICK_HEAD_LEN = 4

type Stick struct {
	conn    io.ReadWriteCloser
	cache   []byte
	index   int
	stop    chan struct{}

	read    int
	write   int
}

func NewStick(rw io.ReadWriteCloser) *Stick {
	s := &Stick{cache: make([]byte, STICK_MAX_BUFF_LEN), conn: rw, stop: make(chan struct{}, 10)}
	return s
}

func (s *Stick)Stop() chan struct{} {
	return s.stop
}

func (s *Stick)Close()  {
	s.conn.Close()
	s.stop <- struct{}{}
}

func (s *Stick)Read() ([]byte, error) {
	for  {
		if s.index >= STICK_MAX_BUFF_LEN {
			return nil, fmt.Errorf("stick buffer full")
		}
		cnt, err := s.conn.Read(s.cache[s.index:])
		if cnt > 0 {
			s.read += cnt

			s.index += cnt
			if s.index < STICK_HEAD_LEN {
				continue
			}
			length := StickHeaderDecoder(s.cache)
			if s.index < (int(length) + STICK_HEAD_LEN) {
				continue
			}
			body := make([]byte, length)
			lastBegin := STICK_HEAD_LEN + int(length)
			copy(body, s.cache[STICK_HEAD_LEN:lastBegin])

			remainLen := s.index - lastBegin
			copy(s.cache, s.cache[lastBegin: lastBegin + remainLen])
			s.index = remainLen

			return body, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func (s *Stick)Write(in []byte) error {
	length := len(in)
	s.write += length

	if length > (STICK_MAX_BUFF_LEN - STICK_HEAD_LEN) {
		return fmt.Errorf("stick buffer not not enough")
	}
	body := make([]byte, length +STICK_HEAD_LEN)
	copy(body[:STICK_HEAD_LEN], StickHeaderCoder(uint32(length)))
	copy(body[STICK_HEAD_LEN:], in)
	return fullWrite(s.conn, body)
}

func StickHeaderCoder(length uint32) []byte {
	var output [STICK_HEAD_LEN]byte
	binary.BigEndian.PutUint32(output[:], length)
	return output[:]
}

func StickHeaderDecoder(in []byte) uint32 {
	return binary.BigEndian.Uint32(in)
}

func fullWrite(conn io.Writer, buf []byte) error {
	var sendcnt int
	for {
		cnt, err := conn.Write(buf[sendcnt:])
		if err != nil {
			return err
		}
		if cnt + sendcnt >= len(buf) {
			return nil
		}
		sendcnt += cnt
	}
}
