package datalayer

import (
	"errors"
	"fmt"
)

type Datalink struct {
	// srcmac represent MAC address
	SrcMac string

	//DstMAC reprsent destination MAC address
	DStMAC string

	//VLAN represent VLAN value
	Vlan int

	// EtherType represent upper layer type value
	EtherType uint16
}

// Packet represents layer 2,3,4 available info
type Packet struct {
	L2   Datalink
	L3   interface{}
	L4   interface{}
	data []byte
}

const (
	// Ether Type is address Resolution Protocol Ether Type value
	EtherTpeARP = 0x0806

	//EtherTypeIPv4 is Internet Protocol version 4 EtherType value
	EtherTypeIPv4 = 0x0800

	// EtherTypeIPv6 is Internet Protocol Version 6 EtherType value
	EtherTypeIPv6 = 0x86DD

	// EtherTypeLACP is Link Aggregation Control Protocol EtherType value
	EtherTypeLACP = 0x8809

	// EtherTypeIEEE8021Q is VLAN-tagged frame (IEEE 802.1Q) EtherType value
	EtherTypeIEEE8021Q = 0x8100
)

var (
	errShortEthernetHeaderLEngth = errors.New("the ethernet header is too small")
)

func (p *Packet) decodeEthernet() error {
	var (
		d   Datalink
		err error
	)

	if len(p.data) < 14 {
		return errShortEthernetHeaderLEngth
	}

	d, err = decodeIEEE802(p.data)
	if err != nil {
		return err
	}

	if d.EtherType == EtherTypeIEEE8021Q {
		vlan := int(p.data[14])<<8 | int(p.data[15])
		p.data[12], p.data[13] = p.data[16], p.data[17]
		p.data = append(p.data[:14], p.data[18:]...)

		d, err = decodeIEEE802(p.data)
		if err != nil {
			return err
		}
		d.Vlan = vlan
	}
	p.L2 = d
	p.data = p.data[14:]

	return nil
}

func decodeIEEE802(b []byte) (Datalink, error) {
	var d Datalink

	if len(b) < 14 {
		return d, errShortEthernetHeaderLEngth
	}
	d.EtherType = uint16(b[13]) | uint16(b[12])<<8

	hwAddrFmt := "%0.2x:%0.2x:%0.2x:%0.2x:%0.2x:%0.2x"

	if d.EtherType != EtherTypeIEEE8021Q {
		d.DStMAC = fmt.Sprintf(hwAddrFmt, b[0], b[1], b[2], b[3], b[4], b[5])
		d.SrcMac = fmt.Sprintf(hwAddrFmt, b[6], b[7], b[8], b[9], b[10], b[11])
	}
	return d, nil
}
