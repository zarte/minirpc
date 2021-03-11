package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"context"
	"google.golang.org/grpc"
	"minirpc/rpc/interf"
	"github.com/hashicorp/consul/api"
	"time"
)

func main() {
	var lastIndex uint64
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500" //consul server

	client, err := api.NewClient(config)
	if err != nil {
		fmt.Println("api new client is failed, err:", err)
		return
	}
	services, metainfo, err := client.Health().Service("sendmailrpc", "v1.0", true, &api.QueryOptions{
		WaitIndex: lastIndex, // 同步点，这个调用将一直阻塞，直到有新的更新
	})
	if err != nil {
		fmt.Println("error Consul: %v", err)
	}
	lastIndex = metainfo.LastIndex

	addrs := map[string]struct{}{}
	if len(services)<=0{
		fmt.Println("no find rpc server")
	}
	for _, service := range services {
		fmt.Println("service.Service.Address:", service.Service.Address, "service.Service.Port:", service.Service.Port)
		addrs[net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))] = struct{}{}
		testsendmail(service.Service.Address,strconv.Itoa(service.Service.Port))
	}
}

func testsendmail(host string,port string)  {
	//建立链接
	conn, err := grpc.Dial(host+":" +port, grpc.WithInsecure())
	if err != nil {
		log.Fatal("did not connect", err)
	}
	defer conn.Close()
	emailerClient := rpc.NewEmailerClient(conn)
	//设定请求超时时间 3s
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	//UserIndex 请求
	res, err := emailerClient.SendMail(ctx, &rpc.SendRequest{
		Touser:     "285568281@qq.com",
		Title: "测试标题",
		Content: "测试内容",
	})

	if err != nil {
		log.Printf("send fail: %v", err)
	}

	if 0 == res.Err {
		fmt.Println("发送成功")
		fmt.Println(res.Msg)
	} else {
		fmt.Println("发送失败",res.Msg)
	}

}