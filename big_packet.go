package gnet

import (
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"unsafe"
)

const (
	// 大包的包头长度
	DefaultBigPacketHeaderSize = int(unsafe.Sizeof(BigPacketHeader{}))
	// 数据包长度限制(4G)
	MaxBigPacketDataSize = 0xFFFFFFFF
)

// 大包包头
// 大包:包长度可能超过16M
type BigPacketHeader struct {
	len     uint32
	command uint16
	flags   uint16
}

func NewBigPacketHeader(len uint32, command, flags uint16) *BigPacketHeader {
	return &BigPacketHeader{
		len:     len,
		command: command,
		flags:   flags,
	}
}

// 包体长度,不包含包头的长度
// [0,0xFFFFFFFF]
func (this *BigPacketHeader) Len() uint32 {
	return this.len
}

// 消息号
func (this *BigPacketHeader) Command() uint16 {
	return this.command
}

// 标记
func (this *BigPacketHeader) Flags() uint16 {
	return this.flags
}

// 从字节流读取数据,len(messageHeaderData)>=MessageHeaderSize
// 使用小端字节序
func (this *BigPacketHeader) ReadFrom(messageHeaderData []byte) {
	this.len = binary.LittleEndian.Uint32(messageHeaderData)
	this.command = binary.LittleEndian.Uint16(messageHeaderData[4:])
	this.flags = binary.LittleEndian.Uint16(messageHeaderData[6:])
}

// 写入字节流,使用小端字节序
func (this *BigPacketHeader) WriteTo(messageHeaderData []byte) {
	binary.LittleEndian.PutUint32(messageHeaderData, this.len)
	binary.LittleEndian.PutUint16(messageHeaderData[4:], this.command)
	binary.LittleEndian.PutUint16(messageHeaderData[6:], this.flags)
}

// 包含一个消息号和[]byte的数据包
type BigDataPacket struct {
	command uint16
	data []byte
}

func NewBigDataPacket(command uint16, data []byte) *BigDataPacket {
	return &BigDataPacket{
		command: command,
		data: data,
	}
}

func (this *BigDataPacket) Command() PacketCommand {
	return PacketCommand(this.command)
}

func (this *BigDataPacket) Message() proto.Message {
	return nil
}

func (this *BigDataPacket) GetStreamData() []byte {
	return this.data
}

// deep copy
func (this *BigDataPacket) Clone() Packet {
	newPacket := &BigDataPacket{data: make([]byte,len(this.data))}
	newPacket.command = this.command
	copy(newPacket.data, this.data)
	return newPacket
}