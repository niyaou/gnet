package main

import (
	"fmt"
	"github.com/gnet"
	"sync"
	"time"
)

func main() {
	netMgr := gnet.GetNetMgr()
	connectionConfig := gnet.ConnectionConfig{
		SendPacketCacheSize: 100,
		RecvTimeout: 5,
		WriteTimeout: 1,
	}
	listenAddress := "127.0.0.1:10002"
	codec := gnet.NewXorCodec([]byte{0,1,2,3,4,5,6})
	netMgr.NewListener(listenAddress, connectionConfig, codec, &EchoServerHandler{}, &EchoListenerHandler{})
	time.Sleep(time.Second)

	netMgr.NewConnector(listenAddress, connectionConfig, codec, &EchoClientHandler{})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	exitTimer := time.NewTimer(10*time.Second)
	select {
	case <-exitTimer.C:
		gnet.LogDebug("test timeout")
		wg.Done()
	}
	wg.Wait()
	netMgr.Shutdown()
}

// 监听接口
type EchoListenerHandler struct {
	
}

func (e EchoListenerHandler) OnConnectionConnected(connection gnet.Connection) {
	gnet.LogDebug(fmt.Sprintf("OnConnectionConnected %v", connection.GetConnectionId()))
}

func (e EchoListenerHandler) OnConnectionDisconnect(connection gnet.Connection) {
	gnet.LogDebug(fmt.Sprintf("OnConnectionDisconnect %v", connection.GetConnectionId()))
}

// 服务端监听到的连接接口
type EchoServerHandler struct {
}

func (e *EchoServerHandler) OnConnected(connection gnet.Connection, success bool) {
	gnet.LogDebug(fmt.Sprintf("Server OnConnected %v %v", connection.GetConnectionId(), success))
	if success {
		// 开一个协程,服务器自动给客户端发消息
		go func() {
			autoSendTimer := time.NewTimer(time.Second)
			serialId := 0
			for connection.IsConnected() {
				select {
				case <-autoSendTimer.C:
					serialId++
					packet := gnet.NewPacket([]byte(fmt.Sprintf("hello client %v", serialId)))
					connection.Send(packet)
					autoSendTimer.Reset(time.Second)
				}
			}
		}()
	}
}

func (e *EchoServerHandler) OnDisconnected(connection gnet.Connection, ) {
	gnet.LogDebug(fmt.Sprintf("Server OnDisconnected %v", connection.GetConnectionId()))
}

func (e *EchoServerHandler) OnRecvPacket(connection gnet.Connection, packet *gnet.Packet) {
	gnet.LogDebug(fmt.Sprintf("Server OnRecvPacket %v: %v", connection.GetConnectionId(), string(packet.GetData())))
}


// 客户端连接接口
type EchoClientHandler struct {
	echoCount int
}

func (e *EchoClientHandler) OnConnected(connection gnet.Connection, success bool) {
	gnet.LogDebug(fmt.Sprintf("Client OnConnected %v %v", connection.GetConnectionId(), success))
}

func (e *EchoClientHandler) OnDisconnected(connection gnet.Connection, ) {
	gnet.LogDebug(fmt.Sprintf("Client OnDisconnected %v", connection.GetConnectionId()))
}

func (e *EchoClientHandler) OnRecvPacket(connection gnet.Connection, packet *gnet.Packet) {
	gnet.LogDebug(fmt.Sprintf("Client OnRecvPacket %v: %v", connection.GetConnectionId(), string(packet.GetData())))
	e.echoCount++
	echoPacket := gnet.NewPacket([]byte(fmt.Sprintf("hello server %v", e.echoCount)))
	connection.Send(echoPacket)
}
