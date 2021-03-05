# emailerrpc
send email service. 发送邮件微服务程序

## 运行
go run emailserver.go -euser xxx -epasswd xxxx   
默认rpc端口1234 
## 参数说明  
consul: consul服务发现注册地址,示例127.0.0.1:8500  
h: 服务外网ip示例127.0.0.1  
rpcport: rpc服务端口示例1234    
checkport: consul健康检查端口示例1233  

邮箱配置   
ehost:邮局地址,示例smtp.qq.co  
eport:邮局端口,示例465  
euser:发件人,示例test@qq.com,必填  
epasswd:发件密码，必填  

