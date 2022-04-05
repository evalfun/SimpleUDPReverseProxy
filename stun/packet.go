// Copyright 2016 Cong Ding
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stun

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
)

type packet struct {
	Types      uint16
	Length     uint16
	TransID    []byte // 4 bytes magic cookie + 12 bytes transaction id
	Attributes []Attribute
}

func NewPacket() (*packet, error) {
	v := new(packet)
	v.TransID = make([]byte, 16)
	binary.BigEndian.PutUint32(v.TransID[:4], magicCookie)
	_, err := rand.Read(v.TransID[4:])
	if err != nil {
		return nil, err
	}
	v.Attributes = make([]Attribute, 0, 10)
	v.Length = 0
	return v, nil
}

func NewPacketFromBytes(packetBytes []byte) (*packet, error) {
	if len(packetBytes) < 20 {
		return nil, errors.New("Received data Length too short.")
	}
	if len(packetBytes) > math.MaxUint16+20 {
		return nil, errors.New("Received data Length too long.")
	}
	pkt := new(packet)
	pkt.Types = binary.BigEndian.Uint16(packetBytes[0:2])
	pkt.Length = binary.BigEndian.Uint16(packetBytes[2:4])
	pkt.TransID = packetBytes[4:20]
	pkt.Attributes = make([]Attribute, 0, 10)
	packetBytes = packetBytes[20:]
	for pos := uint16(0); pos+4 < uint16(len(packetBytes)); {
		types := binary.BigEndian.Uint16(packetBytes[pos : pos+2])
		Length := binary.BigEndian.Uint16(packetBytes[pos+2 : pos+4])
		end := pos + 4 + Length
		if end < pos+4 || end > uint16(len(packetBytes)) {
			return nil, errors.New("Received data format mismatch.")
		}
		value := packetBytes[pos+4 : end]
		Attribute := NewAttribute(types, value)
		pkt.AddAttribute(*Attribute)
		pos += align(Length) + 4
	}
	return pkt, nil
}

func (v *packet) AddAttribute(a Attribute) {
	v.Attributes = append(v.Attributes, a)
	v.Length += align(a.length) + 4
}

func (v *packet) Bytes() []byte {
	packetBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(packetBytes[0:2], v.Types)
	binary.BigEndian.PutUint16(packetBytes[2:4], v.Length)
	packetBytes = append(packetBytes, v.TransID...)
	for _, a := range v.Attributes {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, a.types)
		packetBytes = append(packetBytes, buf...)
		binary.BigEndian.PutUint16(buf, a.length)
		packetBytes = append(packetBytes, buf...)
		packetBytes = append(packetBytes, a.value...)
	}
	return packetBytes
}

func (v *packet) GetSourceAddr() *Host {
	return v.GetRawAddr(AttributeSourceAddress)
}

func (v *packet) GetMappedAddr() *Host {
	return v.GetRawAddr(AttributeMappedAddress)
}

func (v *packet) GetChangedAddr() *Host {
	return v.GetRawAddr(AttributeChangedAddress)
}

func (v *packet) GetOtherAddr() *Host {
	return v.GetRawAddr(AttributeOtherAddress)
}

func (v *packet) GetRawAddr(Attribute uint16) *Host {
	for _, a := range v.Attributes {
		if a.types == Attribute {
			return a.rawAddr()
		}
	}
	return nil
}

func (v *packet) GetXorMappedAddr() *Host {
	addr := v.GetXorAddr(AttributeXorMappedAddress)
	if addr == nil {
		addr = v.GetXorAddr(AttributeXorMappedAddressExp)
	}
	return addr
}

func (v *packet) GetXorAddr(Attribute uint16) *Host {
	for _, a := range v.Attributes {
		if a.types == Attribute {
			return a.xorAddr(v.TransID)
		}
	}
	return nil
}
