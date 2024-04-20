package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort uint
	Name       string //客户端名称
	conn       net.Conn
	mod        int //1:公聊模式 2:私聊模式 3:更改用户名 0:退出
}

// 菜单
func (c *Client) Menu() bool {

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.查询在线用户")
	fmt.Println("0.退出")

	fmt.Println("请选择(0-3):")
	fmt.Scanln(&c.mod)

	if c.mod >= 0 && c.mod <= 4 {
		return true
	}

	fmt.Println(">>>>请输入正确的选项<<<<")
	return false
}

func (c *Client) Run() {
	for c.mod != 0 { //0表示退出
		for !c.Menu() { //返回ture才执行菜单业务
		}
		switch c.mod {
		case 1:
			//公聊模式
			fmt.Println("公聊模式....")
			c.publicChat()
			c.mod = 999
			break
		case 2:
			//私聊模式
			fmt.Println("私聊模式....")
			c.praviteChat()
			break
		case 3:
			//更新用户名
			fmt.Println("更新用户名....")
			c.changeUserName()
			break
		case 4:
			//查询在线用户
			c.showOnlineUser()
		}
	}
}

// 公聊
func (c *Client) publicChat() {
	for {
		var msg string

		fmt.Println(">>>>>请输入要发送的公聊消息,输入exit可以退出<<<<<")
		fmt.Scan(&msg)
		if msg == "exit" {
			break
		}
		fmt.Println(len(msg), msg == "exit")
		c.conn.Write([]byte(msg + "\r\n"))
		msg = ""
	}
}

// 私聊
func (c *Client) praviteChat() {
	for {
		var msg string
		fmt.Println(">>>>>请输入要发送的私公聊消息<<<<<")
		fmt.Println(">>>>>格式为: to|对方用户名|消息内容<<<<<")
		fmt.Scan(&msg)
		if msg == "exit" {
			break
		}
		c.conn.Write([]byte(msg + "\r\n"))
	}
}

// 显示在线用户
func (c *Client) showOnlineUser() {
	_, err := c.conn.Write([]byte("who\r\n"))
	if nil != err {
		fmt.Println("查询在线用户失败!", err)
		return
	}
}

// 修改用户名
func (c *Client) changeUserName() {
	var input string
	fmt.Println(">>>>请输入新用户名<<<<")
	fmt.Scan(&input)
	if input == "" {
		fmt.Println("用户名不能为空!")
		return
	}
	msg := "rename|" + input + "\r\n"
	_, err := c.conn.Write([]byte(msg))
	if nil != err {
		fmt.Println("修改用户名失败!", err)
		return
	}
}

// 处理server回写消息
func (c *Client) DealResponse() {
	//
	io.Copy(os.Stdout, c.conn)
}

// NewClient 创建客户端
func NewClient(serverIp string, serverPort uint) *Client {
	//创建客户端对象
	server := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		mod:        999,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if nil != err {
		fmt.Println("连接服务器失败:", err)
		return nil
	}
	server.conn = conn
	server.Name = conn.RemoteAddr().String()

	//返回对象
	return server
}

var serverIp string
var serverPort uint

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.UintVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	//解析命令行参数
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if nil == client {
		fmt.Println("连接服务器失败....")
		return
	}

	//消息处理
	go client.DealResponse()

	fmt.Println(">>>>>>>连接服务器成功<<<<<<<")

	client.Run()
}
