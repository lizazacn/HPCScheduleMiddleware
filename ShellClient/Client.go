package ShellClient

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"strings"
	"sync"
	"time"
)

// Cli 客户端信息结构体
type Cli struct {
	host       string           // 服务器地址
	port       int              // 服务器端口号
	user       string           // 登录账号
	auth       []ssh.AuthMethod // 密码
	sshClient  *ssh.Client      // 与服务端的连接
	SshSession *ssh.Session     // SSH长连接客户端
	Pipc       *DataChan        // 数据流通道
	UploadTime time.Time        // 更新时间
	ConnStatus bool				// 是否处于连接状态
}

// DataChan 数据通道结构体
type DataChan struct {
	Input    chan string // 数据输入管道
	Output   chan string // 数据输出管道
	OutError chan string // 错误输出管道
	sync     sync.Mutex
}

// NewClient 初始化Cli信息
func NewClient(host string, port int, user string, password string, auth ssh.AuthMethod) *Cli {
	c := new(Cli) // 初始化一个客户端连接
	c.Pipc = new(DataChan)
	c.UploadTime = time.Now()
	//指定输入、输出、错误缓冲区大小
	c.Pipc.Input = make(chan string, 1)
	c.Pipc.Output = make(chan string, 10)
	c.Pipc.OutError = make(chan string, 10)
	c.host = host
	if port <= 0 {
		c.port = 22
	} else {
		c.port = port
	}
	c.user = user
	if password == "" {
		c.auth = append(c.auth, auth)
	} else {
		c.auth = append(c.auth, ssh.Password(password))
	}
	return c
}

// Connect 建立连接
func (c *Cli) Connect() error {
	//建立一个客户端连接
	cliConfig := &ssh.ClientConfig{
		User:            c.user,
		Auth:            c.auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}
	sshClient, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", c.host, c.port), // 拼接连接地址
		cliConfig,
	)
	if err != nil {
		log.Printf("建立ssh连接异常：%v", err)
		return err
	}

	c.sshClient = sshClient
	return nil
}

// Session 建立长连接
func (c *Cli) Session() error {
	var err error
	if c.sshClient == nil {
		if err = c.Connect(); err != nil {
			log.Printf("建立Connect连接异常：%v", err)
			return err
		}
	}
	c.SshSession, err = c.sshClient.NewSession()
	if err != nil {
		log.Printf("建立长连接异常：%v", err)
		return err
	}
	//keys := Util.GetMapKeys(c.SshSession)
	//if len(keys) == 0{
	//	c.SshSession[1] = Session
	//} else {
	//	key := keys[len(keys)-1] + 1
	//	c.SshSession[key] = Session
	//}

	return err
}

// Std 将输入输出改到自定义IO流
func (c *Cli) Std() {
	if c.SshSession == nil {
		err := c.Session()
		if err != nil {
			log.Printf(err.Error())
			return
		}
	}
	session := c.SshSession

	// 将ssh返回数据写入缓冲区
	Reader, err := session.StdoutPipe()
	if err != nil {
		log.Printf("指定输出缓冲区异常：%v", err)
		return
	}

	// 设置进程退出标志
	var quitGo = make(chan bool, 3)
	// 将要发送的数据写入缓冲区
	Write, err := session.StdinPipe()
	if err != nil {
		log.Printf("指定输入缓冲区异常：%v", err)
		return
	}

	Err, err := session.StderrPipe()
	if err != nil {
		log.Printf("指定错误输出缓冲区异常：%v", err)
		return
	}

	scanner := bufio.NewScanner(Reader)
	errout := bufio.NewScanner(Err)

	// 执行命令返回值处理
	go func(s chan bool) {
		//file, _ := os.Create("res.log")
		if c.Pipc.Output == nil {
			c.Pipc.Output = make(chan string, 10)
		}
		var outStatus = (*scanner).Scan()
		for outStatus {
			c.Pipc.Output <- (*scanner).Text()
			select {
			case <-s:
				outStatus = false
				break
			default:
				outStatus = (*scanner).Scan()
				break
			}
		}
		log.Println("用户长连接输出通道正常结束")
	}(quitGo)

	// 输入流
	go func() {
		ok := true
		for ok {
			data, _ := <-c.Pipc.Input
			if strings.Contains(data, "closeTerminal") {
				quitGo <- true
				quitGo <- true
				quitGo <- true
				ok = false
				break
			}
			_, err := Write.Write([]byte(data))
			if err != nil {
				log.Printf("向输入缓冲区写入数据异常：%v", err)
				return
			}
		}
		log.Println("用户长连接输入通道正常关闭！")
	}()

	// 错误流
	go func(s chan bool) {
		if c.Pipc.OutError == nil {
			c.Pipc.OutError = make(chan string, 10)
		}
		scanStu := true
		for scanStu {
			c.Pipc.OutError <- (*errout).Text()
			select {
			case <-s:
				scanStu = false
				break
			default:
				scanStu = (*errout).Scan()
				break
			}
		}
		log.Println("用户长连接错误输出通道正确结束")
	}(quitGo)

	// 监听退出当前进程
	go func(ch chan bool) {
		for true {
			if <-ch {
				break
			}
		}
		log.Println("用户当前长连接进程正确结束")
	}(quitGo)
	err = session.Shell()
	if err != nil {
		log.Printf("Job:Session启动shell连接异常:%v", err)
	}
	err = session.Wait()
	if err != nil {
		log.Printf("Job:Session阻塞shell连接异常:%v", err)
	}
}

// Send 发送消息方法
func (c *Cli) Send(str string) (string, string) {
	if c.Pipc.Input == nil {
		c.Pipc.Input = make(chan string, 1)
	}
	//c.Pipc.Input <- fmt.Sprintf("%s \n echo %s \n", str, user)
	str = strings.Replace(str, " ", " ", -1)
	//str = strings.Replace(str, "&", "", -1)
	indata := "\n echo 'QdzG4w9ugXF5jhFVVX5AesBSfqLNOKix1HLJ' \n echo 'QdzG4w9ugXF5jhFVVX5AesBSfqLNOKix1HLJ' > /dev/stderr \n"
	split := strings.Split(str, "\n")
	for _, val := range split {
		c.Pipc.Input <- val + "\n"
	}
	c.Pipc.Input <- indata
	errStatus := true
	outStatus := true
	var outstr, errstr string
	if c.Pipc.Output != nil {
		for errStatus || outStatus {
			if str == "closeTerminal" {
				break
			}
			if len(c.Pipc.Output) > 1 {
				rstr, _ := <-c.Pipc.Output
				outstr += rstr + "\n"
			} else if len(c.Pipc.Output) == 1 {
				rstr, _ := <-c.Pipc.Output
				if rstr == "QdzG4w9ugXF5jhFVVX5AesBSfqLNOKix1HLJ" && len(c.Pipc.Output) <= 0 {
					outStatus = false
					continue
				}
				outstr += rstr + "\n"
				log.Println("输出数据读取成功！")
			}
			if len(c.Pipc.OutError) > 1 {
				rstr, _ := <-c.Pipc.OutError
				errstr += rstr + "\n"
			} else if len(c.Pipc.OutError) == 1 {
				rstr, _ := <-c.Pipc.OutError
				if rstr == "QdzG4w9ugXF5jhFVVX5AesBSfqLNOKix1HLJ" && len(c.Pipc.OutError) <= 0 {
					errStatus = false
					continue
				}
				errstr += rstr + "\n"
				log.Println("错误数据读取成功！")
			}
		}
	}
	return outstr, errstr
}

// Close 关闭数据管道
func (c Cli) Close() {
	if c.SshSession != nil{
		c.SshSession.Close()
	}
	c.sshClient.Close()
	close(c.Pipc.Output)
	close(c.Pipc.OutError)
	close(c.Pipc.Input)
}
