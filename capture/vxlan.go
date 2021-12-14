package capture

import (
	"errors"
	"fmt"
	"net"
	"time"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const VxLanPacketSize = 1526 //vxlan 8 B + ethernet II 1518 B

type vxlanHandle struct {
	connexion     *net.UDPConn
	packetChannel chan gopacket.Packet
}

func newVXLANHandler() (*vxlanHandle, error) {
	addr := net.UDPAddr{
		Port: 4789,
		IP:   net.ParseIP("0.0.0.0"),
	}
	vxlanHandle := &vxlanHandle{}
	con, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	vxlanHandle.connexion = con
	vxlanHandle.packetChannel = make(chan gopacket.Packet, 1000)
	go vxlanHandle.reader()

	return vxlanHandle, nil
}

func (v *vxlanHandle) reader() {
	for {
		inputBytes := make([]byte, VxLanPacketSize)
		length, _, err := v.connexion.ReadFromUDP(inputBytes)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		packet := gopacket.NewPacket(inputBytes[:length], layers.LayerTypeVXLAN, gopacket.NoCopy)
		ci := packet.Metadata()
		ci.Timestamp = time.Now()
		ci.CaptureLength = length
		ci.Length = length
		v.packetChannel <- packet
	}
}

func (v *vxlanHandle) ReadPacketData() ([]byte, gopacket.CaptureInfo, error) {
	packet := <-v.packetChannel
	bytes := packet.Layer(layers.LayerTypeVXLAN).LayerPayload()

	return bytes, packet.Metadata().CaptureInfo, nil
}



func (v *vxlanHandle) Close() error {
	if v.connexion != nil {
		return v.connexion.Close()
	}
	return nil
}
