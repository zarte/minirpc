package service

import (
	"fmt"
	"log"
	rpc "minirpc/rpc/interf"
	"google.golang.org/grpc"
	"time"
	"context"
	"errors"
)

type EmailContent struct {
	Touser string
	Title string
	Content string
}
func Sendmail(host string,port string, content EmailContent) error  {
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
		Touser:  content.Touser,
		Title: content.Title,
		Content: content.Content,
	})

	if err != nil {
		log.Printf("send fail: %v", err)
	}

	if 0 == res.Err {
		fmt.Println("发送成功")
		fmt.Println(res.Msg)
		return nil
	} else {
		return errors.New("res.Msg")
	}

}
