/*
 * Copyright (C) 2020-2022, IrineSistiana
 *
 * This file is part of mosdns.
 *
 * mosdns is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * mosdns is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package dnsutils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/UFR6cRY9xufLKtx2idrc/mosdns/main/pkg/pool"
	"github.com/miekg/dns"
	"io"
)

var (
	errZeroLenMsg = errors.New("zero length msg")
)

// ReadRawMsgFromTCP reads msg from c in RFC 1035 format (msg is prefixed
// with a two byte length field).
// n represents how many bytes are read from c.
// The returned the []byte should be released by pool.ReleaseBuf.
func ReadRawMsgFromTCP(c io.Reader) ([]byte, int, error) {
	n := 0
	h := pool.GetBuf(2)
	defer pool.ReleaseBuf(h)
	nh, err := io.ReadFull(c, h)
	n += nh
	if err != nil {
		return nil, n, err
	}

	// dns length
	length := binary.BigEndian.Uint16(h)
	if length == 0 {
		return nil, 0, errZeroLenMsg
	}

	buf := pool.GetBuf(int(length))
	nm, err := io.ReadFull(c, buf)
	n += nm
	if err != nil {
		pool.ReleaseBuf(buf)
		return nil, n, err
	}
	return buf, n, nil
}

// ReadMsgFromTCP reads msg from c in RFC 1035 format (msg is prefixed
// with a two byte length field).
// n represents how many bytes are read from c.
func ReadMsgFromTCP(c io.Reader) (*dns.Msg, int, error) {
	b, n, err := ReadRawMsgFromTCP(c)
	if err != nil {
		return nil, 0, err
	}
	defer pool.ReleaseBuf(b)

	m, err := unpackMsgWithDetailedErr(b)
	return m, n, err
}

// WriteMsgToTCP packs and writes m to c in RFC 1035 format.
// n represents how many bytes are written to c.
func WriteMsgToTCP(c io.Writer, m *dns.Msg) (n int, err error) {
	mRaw, buf, err := pool.PackBuffer(m)
	if err != nil {
		return 0, err
	}
	defer pool.ReleaseBuf(buf)
	return WriteRawMsgToTCP(c, mRaw)
}

// WriteRawMsgToTCP See WriteMsgToTCP
func WriteRawMsgToTCP(c io.Writer, b []byte) (n int, err error) {
	if len(b) > dns.MaxMsgSize {
		return 0, fmt.Errorf("payload length %d is greater than dns max msg size", len(b))
	}

	buf := pool.GetBuf(len(b) + 2)
	defer pool.ReleaseBuf(buf)

	binary.BigEndian.PutUint16(buf[:2], uint16(len(b)))
	copy(buf[2:], b)
	return c.Write(buf)
}

func WriteMsgToUDP(c io.Writer, m *dns.Msg) (int, error) {
	b, buf, err := pool.PackBuffer(m)
	if err != nil {
		return 0, err
	}
	defer pool.ReleaseBuf(buf)

	return c.Write(b)
}

func ReadMsgFromUDP(c io.Reader, bufSize int) (*dns.Msg, int, error) {
	if bufSize < dns.MinMsgSize {
		bufSize = dns.MinMsgSize
	}

	b := pool.GetBuf(bufSize)
	defer pool.ReleaseBuf(b)
	n, err := c.Read(b)
	if err != nil {
		return nil, n, err
	}

	m, err := unpackMsgWithDetailedErr(b[:n])
	return m, n, err
}

func unpackMsgWithDetailedErr(b []byte) (*dns.Msg, error) {
	m := new(dns.Msg)
	if err := m.Unpack(b); err != nil {
		return nil, fmt.Errorf("failed to unpack msg [%x], %w", b, err)
	}
	return m, nil
}
