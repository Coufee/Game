package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tool/common"
	l4g "tool/log4go"
	"tool/zebra"

	//"io/ioutil"
	_ "os"
	"path/filepath"
)

var g_config = new(xmlConfig)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		l4g.Fine("%v", err)
		//l4g.Fatal("%v", err)
	}
	fmt.Println(dir)

	// if len(os.Args) != 2 {
	// 	fmt.Println("Please input ", os.Args[0], "config file")
	// }

	dirs := "server.xml"
	if err := common.LoadConfig(dirs, g_config); err != nil {
		panic(fmt.Sprintf("load config %v fail %v", os.Args[1], err))
	}

	// l4g.LoadConfiguration(g_config.Log.File)
	// defer l4g.Close()
	l4g.Fine("cqtest start")
	server_conf := &zebra.Config{}
	server_conf.Address = ":" + fmt.Sprint(g_config.In.Port)
	server_conf.MaxReadMsgSize = int(g_config.In.Max_read_msg_size)
	server_conf.MaxWriteMsgSize = int(g_config.In.Max_write_msg_size)
	server_conf.ReadTimeOut = 600
	server_conf.WriteTimeOut = 600
	server_conf.ReadMsgQueneSize = 1024
	server_conf.WriteMsgQueneSize = 1024
	go zebra.TCPServe(g_cs, server_conf)
	client_conf := &zebra.Config{}
	client_conf.Address = ":" + fmt.Sprint(g_config.Out.Port)
	client_conf.MaxReadMsgSize = int(g_config.Out.Max_read_msg_size)
	client_conf.MaxWriteMsgSize = int(g_config.Out.Max_write_msg_size)
	client_conf.ReadTimeOut = 600
	client_conf.WriteTimeOut = 600
	client_conf.ReadMsgQueneSize = 1024
	client_conf.WriteMsgQueneSize = 1024
	go zebra.TCPServe(g_cs, client_conf)
	g_ctp = NewThreadPool(int(g_config.Out.Broadcast))

	//linux终止进程操作
	signal.Notify(g_signal, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	g_timer.AddTimer(&TenSecondTimer{}, time.Now().Unix(), 10)
	g_timer.AddTimer(&TenMinuteTimer{}, time.Now().Unix(), 4)
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case t := <-ticker.C:
			g_timer.Run(t.Unix(), 0)
		case sig := <-g_signal:
			l4g.Info("Signal: %s", sig.String())
			return
		}
	}

}
