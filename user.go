package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	Conn net.Conn

	server *Server
}

// 创建用户Api
func NewUser(conn net.Conn, s *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		Conn:   conn,
		server: s,
	}

	//启动监听当前userchannel 消息goroutine
	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法, 一旦有消息,就发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.Conn.Write([]byte(msg + "\r\n"))
	}
}

// 用户上线
func (this *User) Login() {
	//用户上线 用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "login")
}

// 用户下线
func (this *User) Logout() {
	//用户上线 用户加入到onlineMap中
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "logout")
}

// 给当前客户端发送消息
func (this *User) SendMsg(msg string) {
	this.Conn.Write([]byte(msg))
}

// 修改用户名方法
func (this *User) ChangeName(newName string) {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name) //先删老数据
	this.server.OnlineMap[newName] = this    //在增新数据
	this.server.mapLock.Unlock()

	this.Name = newName
	this.SendMsg("change name success!, newName:" + newName + "\r\n")
}

// 发送消息广播
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询所有在线用户
		this.SendMsg("------OnlineList------\r\n")
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":is online\r\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
		this.SendMsg("----------------------\r\n")
		return
		//修改用户名
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.TrimPrefix(msg, "rename|")
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("newname exist!!\r\n")
			return
		}
		this.ChangeName(newName)
		return
	} else if len(msg) > 4 && strings.Split(msg, "|")[0] == "to" {
		// 消息协议格式 to|目标|具体消息
		//目标名称
		targetName := strings.Split(msg, "|")[1]

		targerUser, ok := this.server.OnlineMap[targetName]
		if !ok {
			fmt.Println("no target user found!")
			return
		}
		//具体消息内容
		content := strings.Split(msg, "|")[2]
		if content == "" {
			fmt.Println("msg cannot be empty!! ")
			return
		}

		targerUser.SendMsg(fmt.Sprintf("[%s]Tell You:%s\r\n", this.Name, content))

	} else {
		this.server.BroadCast(this, msg)
	}
}
