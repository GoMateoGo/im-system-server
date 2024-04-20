package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户map
	OnlineMap map[string]*User
	// 读写锁
	mapLock sync.RWMutex
	//消息广播channel
	Message chan string
}

// 创建server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 广播消息方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 监听MessageChannel的goroutine,一旦有消息,就发送给全部在线的User
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 业务处理
func (this *Server) Handler(conn net.Conn) {
	//userAddr := conn.RemoteAddr().String()
	//fmt.Printf("客户端%s 连接到服务器\n", userAddr)

	//用户上线
	user := NewUser(conn, this)
	user.Login()

	//isLive channel
	isLevel := make(chan bool)

	//接受客户端发来的消息
	go func() {
		buf := make([]byte, 4096)
		msg := ""
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Logout()
				return
			}

			if nil != err && err != io.EOF {
				fmt.Printf("读取客户端[%s]消息错误:%s", user.Name, err.Error())
				return
			}

			msg += string(buf[:n])

			if strings.HasSuffix(msg, "\r\n") {
				user.DoMessage(strings.TrimSuffix(msg, "\r\n"))
				msg = ""
			}

			//用户的任意消息,都代表为活跃状态
			isLevel <- true
		}
	}()

	for {
		select {
		case <-isLevel:
			//啥都不做
		case <-time.After(300 * time.Second):
			//超时T下线
			user.SendMsg("timeout !!")

			//关闭isLive chan
			close(isLevel)
			//关闭连接
			conn.Close()
			return
		}
	}
}

// 启动服务
func (this Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if nil != err {
		fmt.Println("net.Listen 监听错误 :", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Socket Listen 开启监听%s:%d\n", this.Ip, this.Port)

	//监听消息
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if nil != err {
			fmt.Println("socket listenner 监听错误 :", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}

	//close listen socket
}
