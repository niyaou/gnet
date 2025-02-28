package example

import (
	"context"
	. "github.com/fish-tennis/gnet"
	"github.com/fish-tennis/gnet/example/pb"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestPacketSize(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			logger.Debug("fatal %v", err.(error))
			LogStack()
		}
	}()

	SetLogLevel(InfoLevel)
	// 3秒后触发关闭通知,所有监听<-ctx.Done()的地方会收到通知
	ctx,cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	connectionConfig := ConnectionConfig{
		SendPacketCacheCap: 4,
		SendBufferSize:     16,
		RecvBufferSize:     16,
		MaxPacketSize:      60, // 超出RingBuffer的大小
	}
	listenAddress := "127.0.0.1:10002"
	defaultCodec := NewProtoCodec(nil)
	serverHandler := NewDefaultConnectionHandler(defaultCodec)
	serverHandler.Register(PacketCommand(123), func(connection Connection, packet *ProtoPacket) {
		testMessage := packet.Message().(*pb.TestMessage)
		logger.Info("recv%v:%s",testMessage.I32, testMessage.Name)
	}, func() proto.Message {
		return new (pb.TestMessage)
	})
	if GetNetMgr().NewListener(ctx, listenAddress, connectionConfig, defaultCodec, serverHandler, nil) == nil {
		panic("listen failed")
	}

	clientHandler := NewDefaultConnectionHandler(defaultCodec)
	clientConnector := GetNetMgr().NewConnector(ctx, listenAddress, &connectionConfig, defaultCodec, clientHandler, nil)
	if clientConnector == nil {
		panic("connect failed")
	}

	go func() {
		for i := 0; i < 100; i++ {
			// 发送一个随机大小的数据包,有一些会超出RingBuffer大小
			testMessage := &pb.TestMessage{
				I32:int32(i+1),
			}
			randomLen := 1 + rand.Intn(50)
			for j := 0; j < randomLen; j++ {
				testMessage.Name += strconv.Itoa(j%10)
			}
			logger.Info("send%v:%s",testMessage.I32, testMessage.Name)
			clientConnector.Send(123, testMessage)
		}
	}()

	GetNetMgr().Shutdown(true)
}
