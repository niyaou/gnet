package gnet

// 不使用RingBuffer的编解码
type CodecNoRing struct {
}

// 使用BigPacketHeader
func (this *CodecNoRing) CreatePacketHeader(connection Connection, packet Packet, packetData []byte) PacketHeader {
	if packet == nil {
		return NewBigPacketHeader(uint32(len(packetData)), 0,0)
	}
	return NewBigPacketHeader(uint32(len(packetData)), uint16(packet.Command()),0)
}

func (this *CodecNoRing) PacketHeaderSize() uint32 {
	return uint32(DefaultBigPacketHeaderSize)
}

// 直接返回原包的字节流数据
func (this *CodecNoRing) Encode(connection Connection, packet Packet) []byte {
	return packet.GetStreamData()
}

// data包含了包头
func (this *CodecNoRing) Decode(connection Connection, data []byte) (newPacket Packet, err error) {
	packetHeader := &BigPacketHeader{}
	packetHeader.ReadFrom(data[0:])
	newPacket = NewBigDataPacket(packetHeader.Command(), data[DefaultBigPacketHeaderSize:])
	return
}