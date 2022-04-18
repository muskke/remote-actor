package main

import (
	"fmt"
	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/micro-community/micro/v3/service/logger"
	pb "github.com/muskke/remote-actor/proto"
	"github.com/muskke/remote-actor/utils"
	"strconv"
	"time"
)

const RemoteIP = "127.0.0.1"

func main() {
	// 初始化远程配置
	localIP, _ := utils.GetOutBoundIP()
	system := actor.NewActorSystem()
	remoteConfig := remote.Configure(localIP, 0) //作为client节点，只需配置IP即可
	remoting := remote.NewRemote(system, remoteConfig)
	remoting.Start()

	// 构造本地Actor上下文
	rootContext := system.Root

	// 创建远程事件监听
	props := actor.PropsFromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *pb.Ready:
			logger.Infof("[%+v] ready: %+v", context.Self().RequestId, msg.Sender)
			context.RequestWithCustomSender(context.Sender(), &pb.Done{Sender: context.Self()}, context.Self())
		case *pb.Done:
			logger.Infof("[%+v] done: %+v", context.Self().RequestId, msg.Sender)
			context.RequestWithCustomSender(context.Sender(), &pb.Ready{Sender: context.Self()}, context.Self())
		case *actor.Stop, *actor.Stopped:
			context.RequestWithCustomSender(context.Sender(), &actor.Stopped{}, context.Self()) //通知对方自己已停用
			context.Stop(context.Self())                                                        //释放自己
		default:
			logger.Infof("[%+v] other: %+v", context.Self().RequestId, msg)
		}
		<-time.After(time.Millisecond * 10)
	})

	// 异步并发测试
	ticker := time.NewTicker(time.Millisecond * 100)
	i := 0
	for range ticker.C {
		i += 1
		go func() {
			// 创建本地Actor
			client := rootContext.Spawn(props)

			// 创建远程Actor
			server := actor.NewPID(fmt.Sprintf("%v:%v",RemoteIP,utils.ServerPort), strconv.Itoa(i))
			actor.NewActorSystem()

			// 创建消息通道
			channel := newMessageSenderChannel(rootContext, server)

			// 发送本地事件到远程
			channel <- &pb.Done{
				Sender: client,
			} //事件结构体

			// rootContext.Send(server, &pb.Ready{Sender: client})  //XXX: 该方法发送，没有Sender，属于匿名发送
			rootContext.RequestWithCustomSender(server, &pb.Ready{Sender: client}, client)

			// 10s后释放自己
			time.AfterFunc(time.Second*10, func() {
				rootContext.Stop(client)
			})
		}()
	}

	//// ==============单体测试
	////创建本地Actor
	//client := rootContext.Spawn(props)
	//
	//// 创建远程Actor
	//server := actor.NewPID(fmt.Sprintf("%v:%v",RemoteIP,utils.ServerPort), "actor001")
	//actor.NewActorSystem()
	//
	//// 创建消息通道
	//channel := newMessageSenderChannel(rootContext, server)
	//
	//// 发送本地事件到远程
	//channel <- &pb.Done{
	//	Sender: client,
	//} //事件结构体
	//
	//rootContext.RequestWithCustomSender(server, &pb.Ready{Sender: client}, rootContext.Self())

	_, _ = console.ReadLine()
}

func newMessageSenderChannel(context actor.SenderContext, remotePID *actor.PID) chan<- *pb.Done {
	channel := make(chan *pb.Done)
	go func() {
		for msg := range channel {
			context.RequestWithCustomSender(remotePID, msg, context.Self())
			_ = msg
		}
	}()

	return channel

}
