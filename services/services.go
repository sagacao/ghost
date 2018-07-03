package services

import (
	. "ghost/global"
	"ghost/network"
	"ghost/services/handler"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	closeSig    chan bool
	tcpServer   *network.TCPServer
	GHttpClient *network.HttpClient
)

func init() {
	closeSig = make(chan bool, 1)
}

func registSignal() {
	var stopLock sync.Mutex
	signalChan := make(chan os.Signal, 1)
	go func() {
		//阻塞程序运行，直到收到终止的信号
		<-signalChan
		stopLock.Lock()
		closeSig <- true
		stopLock.Unlock()
		Log.Info("Safe Stop TCP Server ...")
		return
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
}

func safeClose() {
	go func() {
		<-closeSig
		if tcpServer != nil {
			tcpServer.Close()
		}
		os.Exit(1)
		return
	}()
}

func RunTCPServer() {

	registSignal()

	tcpServer = new(network.TCPServer)
	tcpServer.Addr = Config.Tcp.Addr
	tcpServer.MaxConnNum = 100
	tcpServer.PendingWriteNum = 1000
	tcpServer.LenMsgLen = 4
	tcpServer.MaxMsgLen = 4096
	tcpServer.LittleEndian = true
	tcpServer.NewAgent = func(sess *network.ConnSession) network.Agent {
		return &network.AgentInstance{Sess: sess, Worker: handler.Processor}
	}

	if tcpServer != nil {
		tcpServer.Start()
	}

	safeClose()
}

func RunAsycHttpServer() {

	GHttpClient = new(network.HttpClient)
	GHttpClient.PendingWriteNum = 1000

	if GHttpClient != nil {
		GHttpClient.Start()
	}
}
