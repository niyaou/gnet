package example

import (
	"context"
	"fmt"
	. "github.com/fish-tennis/gnet"
	"net"
	"testing"
	"time"
)

// 测试大包
func TestBigPacket(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("fatal %v", err.(error))
			LogStack()
		}
	}()

	SetLogLevel(DebugLevel)
	// 10秒后触发关闭通知,所有监听<-ctx.Done()的地方会收到通知
	ctx,cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	netMgr := GetNetMgr()
	connectionConfig := ConnectionConfig{
		SendPacketCacheCap: 8,
		MaxPacketSize:      1024*1024*64, // 64M
		HeartBeatInterval:  3,
		RecvTimeout:        0,
		WriteTimeout:       0,
	}
	listenAddress := "127.0.0.1:10002"

	serverCodec := &CodecNoRing{}
	serverHandler := &echoBigPacketServerHandler{}
	// 自定义TcpConnection
	if netMgr.NewListenerCustom(ctx, listenAddress, connectionConfig, serverCodec, serverHandler, nil, func(conn net.Conn, config *ConnectionConfig, codec Codec, handler ConnectionHandler) Connection {
		return NewTcpConnectionNoRingAccept(conn, config, codec, handler)
	}) == nil {
		panic("listen failed")
	}
	time.Sleep(time.Second)

	clientCodec := &CodecNoRing{}
	clientHandler := &echoBigPacketClientHandler{}
	// 自定义TcpConnection
	if netMgr.NewConnectorCustom(ctx, listenAddress, &connectionConfig, clientCodec, clientHandler, nil, func(config *ConnectionConfig, codec Codec, handler ConnectionHandler) Connection {
		return NewTcpConnectionNoRing(config, codec, handler)
	}) == nil {
		panic("connect failed")
	}

	netMgr.Shutdown(true)
}

// 服务端监听到的连接接口
type echoBigPacketServerHandler struct {
}

func (e *echoBigPacketServerHandler) OnConnected(connection Connection, success bool) {
	logger.Debug(fmt.Sprintf("Server OnConnected %v %v", connection.GetConnectionId(), success))
	if success {
		// 开一个协程,服务器自动给客户端发消息
		serialId := 0
		packetDataSize := 1024*1024*50
		// 先连发2个数据包
		for i := 0; i < 2; i++ {
			serialId++
			// 模拟一个50M的包
			packetData := make([]byte, packetDataSize, packetDataSize)
			for j := 0; j < len(packetData); j++ {
				packetData[j] = byte(j)
			}
			packet := NewBigDataPacket(2, packetData)
			connection.SendPacket(packet)
		}
		go func() {
			autoSendTimer := time.NewTimer(time.Second)
			for connection.IsConnected() {
				select {
				case <-autoSendTimer.C:
					serialId++
					// 模拟一个50M的包
					packetData := make([]byte, packetDataSize, packetDataSize)
					for j := 0; j < len(packetData); j++ {
						packetData[j] = byte(j)
					}
					packet := NewBigDataPacket(2, packetData)
					connection.SendPacket(packet)
					autoSendTimer.Reset(time.Second)
				}
			}
		}()
	}
}

func (e *echoBigPacketServerHandler) OnDisconnected(connection Connection ) {
	logger.Debug(fmt.Sprintf("Server OnDisconnected %v", connection.GetConnectionId()))
}

func (e *echoBigPacketServerHandler) OnRecvPacket(connection Connection, packet Packet) {
	if len(packet.GetStreamData()) < 100 {
		logger.Debug(fmt.Sprintf("Server OnRecvPacket %v: %v", connection.GetConnectionId(), string(packet.GetStreamData())))
	} else {
		logger.Debug(fmt.Sprintf("Server OnRecvPacket %v: len:%v", connection.GetConnectionId(), len(packet.GetStreamData())))
	}
}

// 服务器不需要发送心跳请求包
func (e *echoBigPacketServerHandler) CreateHeartBeatPacket(connection Connection, ) Packet { return nil }

// 客户端连接接口
type echoBigPacketClientHandler struct {
	echoCount int
}

func (e *echoBigPacketClientHandler) OnConnected(connection Connection, success bool) {
	logger.Debug(fmt.Sprintf("Client OnConnected %v %v", connection.GetConnectionId(), success))
}

func (e *echoBigPacketClientHandler) OnDisconnected(connection Connection ) {
	logger.Debug(fmt.Sprintf("Client OnDisconnected %v", connection.GetConnectionId()))
}

func (e *echoBigPacketClientHandler) OnRecvPacket(connection Connection, packet Packet) {
	if len(packet.GetStreamData()) < 100 {
		logger.Debug(fmt.Sprintf("Client OnRecvPacket %v: %v", connection.GetConnectionId(), string(packet.GetStreamData())))
	} else {
		logger.Debug(fmt.Sprintf("Client OnRecvPacket %v: len:%v", connection.GetConnectionId(), len(packet.GetStreamData())))
	}
	e.echoCount++
	// 模拟一个回复包
	echoPacket := NewBigDataPacket(3, []byte(fmt.Sprintf("hello server %v", e.echoCount)))
	connection.SendPacket(echoPacket)
}

// 客户端定时发送心跳请求包
func (e *echoBigPacketClientHandler) CreateHeartBeatPacket(connection Connection) Packet {
	return NewBigDataPacket(1, []byte("heartbeat"))
}