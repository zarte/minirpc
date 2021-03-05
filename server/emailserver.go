package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-gomail/gomail"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"minirpc/rpc/interf"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"
)
var consulhost string
var selfhost string

var sendmailrpcport int
var checkrpcport int

// consul 服务端会自己发送请求，来进行健康检查
func consulCheck(w http.ResponseWriter, r *http.Request) {
	s := "consulCheck remote:" + r.RemoteAddr + " " + r.URL.String()
	fmt.Println(s)
}

func registerServer() {

	config := consulapi.DefaultConfig()
	config.Address = consulhost
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "sendmailnode1"      // 服务节点的名称
	registration.Name = "sendmailrpc"      // 服务名称
	registration.Port = sendmailrpcport              // 服务端口
	registration.Tags = []string{"v1.0"} // tag，可以为空
	//registration.Address = localIP()      // 服务 IP
	registration.Address = selfhost      // 服务 IP

	checkPort := checkrpcport
	registration.Check = &consulapi.AgentServiceCheck{ // 健康检查
		HTTP:                           fmt.Sprintf("http://%s:%d%s", registration.Address, checkPort, "/check"),
		Timeout:                        "5s",
		Interval:                       "10s",  // 健康检查间隔
		DeregisterCriticalServiceAfter: "60s", //check失败后30秒删除本服务，注销时间，相当于过期时间
		// GRPC:     fmt.Sprintf("%v:%v/%v", IP, r.Port, r.Service),// grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal("register server error : ", err)
	}

	http.HandleFunc("/check", consulCheck)
	http.ListenAndServe(fmt.Sprintf(":%d", checkPort), nil)

}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
type EmailerServer struct {
	User string
	Passwd string
	Host string
	Port int
}
func (u *EmailerServer) SendMail(ctx context.Context, in *rpc.SendRequest) (*rpc.SendResponse, error) {

	if in.Touser==""{
		return &rpc.SendResponse{
			Err: -1,
			Msg: "接收人缺少",
		}, nil
	}
	if in.Title=="" || in.Content==""{
		return &rpc.SendResponse{
			Err: -1,
			Msg: "标题或邮件内容缺少",
		}, nil
	}
	//实现具体邮件发送
	body := in.Content
	body = strings.Replace(body, "\n", "<br/>", -1)
	gomailMessage := gomail.NewMessage()
	gomailMessage.SetHeader("From", u.User)
	gomailMessage.SetHeader("To", in.Touser)
	gomailMessage.SetHeader("Subject", in.Title)
	gomailMessage.SetBody("text/html", in.Content)
	mailer := gomail.NewDialer(u.Host, u.Port,u.User, u.Passwd)
	maxTimes := 3
	i := 0
	sendflag := false
	for i < maxTimes {
		err := mailer.DialAndSend(gomailMessage)
		if err == nil {
			sendflag = true
			break
		}
		i += 1
		time.Sleep(2 * time.Second)
		if i < maxTimes {
			fmt.Println("mail#发送消息失败#%s#消息内容-%s", err.Error(), in.Content)
		}
	}
	if sendflag {
		fmt.Println("发送邮件成功",in.Touser)
		return &rpc.SendResponse{
			Err: 0,
			Msg: "success",
		}, nil
	}else{
		return &rpc.SendResponse{
			Err: -1,
			Msg: "尝试发送"+strconv.Itoa(maxTimes)+"次失败",
		}, nil
	}

}

func starEmailService(host string,port int,user string,passwd string)  {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(sendmailrpcport))
	if err != nil {
		log.Fatal("failed to listen", err)
	}

	//创建rpc服务
	grpcServer := grpc.NewServer()

	//为User服务注册业务实现 将User服务绑定到RPC服务器上
	rpc.RegisterEmailerServer(grpcServer, &EmailerServer{
		User:user,
		Passwd:passwd,
		Host:host,
		Port:port,
	})

	//注册反射服务， 这个服务是CLI使用的， 跟服务本身没有关系
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("faild start rpc,", err)
	}
}


func main() {
	var ehost string
	var eport int
	var euser string
	var epasswd string

	flag.StringVar(&consulhost,"consul","127.0.0.1:8500", "Use -consul <127.0.0.1:8500>")
	flag.StringVar(&selfhost,"h","127.0.0.1", "Use -h <host>")
	flag.IntVar(&sendmailrpcport,"rpcport",1234, "Use -rpcport <rpcport>")
	flag.IntVar(&checkrpcport,"checkport",1233,"Use -checkport <checkport>")
	flag.StringVar(&ehost,"ehost","smtp.qq.com", "Use -ehost <ehost>")
	flag.IntVar(&eport,"eport",465, "Use -eport <eport>")
	flag.StringVar(&euser,"euser","", "Use -euser <euser>")
	flag.StringVar(&epasswd,"epasswd","", "Use -epasswd <epasswd>")
	flag.Parse()
	if epasswd =="" || euser ==""{
		fmt.Println("email user or passwd no set!")
		return
	}
	if consulhost =="" {
		//不使用consul
		fmt.Println("start rpc!")
		starEmailService(ehost,eport,euser,epasswd)
	}else{
		go func() {
			fmt.Println("start rpc!")
			starEmailService(ehost,eport,euser,epasswd)
		}()
		fmt.Println("register consul!")
		registerServer()
	}

}

