package main

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/muskke/remote-actor/server-node/handler"
	"github.com/muskke/remote-actor/utils"
	"strconv"
	"time"

	console "github.com/asynkron/goconsole"
	"github.com/asynkron/protoactor-go/remote"
	"github.com/micro-community/micro/v3/service"
	"github.com/micro-community/micro/v3/service/logger"
	pb "github.com/muskke/remote-actor/proto"
)

func main() {
	srv := &service.Service{}
	srv.Init(
		service.Name("remote-actor"),
		service.Address(fmt.Sprintf("0.0.0.0:%v", utils.ServerPort+1)),
	)

	// Register Handler
	pb.RegisterRemoteActorHandler(srv.Server(), handler.NewHandler())

	go startActorServer()

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatalf("failed to serve: %v", err)
	}
}

func startActorServer() {
	// 开启远程Actor
	system := actor.NewActorSystem()
	//localIP, _ := utils.GetOutBoundIP()
	//remoteConfig := remote.Configure(localIP, utils.ServerPort) // 本节点Actor的远程暴露地址
	remoteConfig := remote.Configure("0.0.0.0", utils.ServerPort) // 本节点Actor的远程暴露地址
	remoting := remote.NewRemote(system, remoteConfig)
	remoting.Start()

	// 创建一个通道，用来传递消息
	channel := make(chan []string)
	go dealMsg(channel)

	// 创建一个接收消息并将它们推送到通道的actor
	props := actor.PropsFromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *pb.Ready:
			logger.Infof("[%+v] Ready: %+v", context.Self().RequestId, msg.Sender)
			context.RequestWithCustomSender(msg.Sender, &pb.Done{Sender: context.Self()}, context.Self())
		case *actor.Started:
			logger.Infof("[%+v] Started %+v", context.Self().RequestId, msg)
		case *actor.Stop, *actor.Stopped:
			logger.Infof("[%+v] Stopped %+v", context.Self().RequestId, msg)
			context.RequestWithCustomSender(context.Sender(), &actor.Stopped{}, context.Self()) //通知对方自己已停用
			context.Stop(context.Self())                                                        //释放自己
		case *pb.Done:
			logger.Infof("[%+v] Done %+v", context.Self().RequestId, msg.Sender)
			context.RequestWithCustomSender(msg.Sender, &pb.Ready{Sender: context.Self()}, context.Self())

		default:
			logger.Errorf("%+v", msg)
		}
		<-time.After(time.Millisecond * 10)
	})

	// ==============================异步并发测试
	i := 0
	ticker := time.NewTicker(time.Millisecond * 100)
	for range ticker.C {
		i += 1
		go func() {
			// 创建一个名叫schedule001的Actor
			pid, err := system.Root.SpawnNamed(props, strconv.Itoa(i))
			logger.Infof("pid:[%+v] - %v", pid, err)

			// 10s后释放自己
			time.AfterFunc(time.Second*10, func() {
				system.Root.Stop(pid)
			})
		}()
	}

	//// ==============================单体测试
	//// 监听actor001
	//pid, err := system.Root.SpawnNamed(props, "actor001")
	//logger.Infof("pid:[%+v] - %v", pid, err)

	// 保活
	console.ReadLine()
}

func dealMsg(ch chan []string) {
	for msg := range ch {
		logger.Infof("%+v", msg)
	}
}
