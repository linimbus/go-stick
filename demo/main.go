package main

import (
	"flag"
	"log"
	"net"
	stick "github.com/lixiangyun/go_stick"
	"time"
)

var (
	HELP   bool
	ADDR   string
	SERVER bool
)

func init()  {
	flag.BoolVar(&HELP, "help", false, "usage help")
	flag.StringVar(&ADDR, "addr", "127.0.0.1:3333", "address for server listen/client connect")
	flag.BoolVar(&SERVER, "server", false, "server/client")
}

var packCnt  int64
var packSize int64

func init()  {
	go func() {
		for  {
			log.Printf("pack cnt %d, size: %d\n", packCnt, packSize)
			time.Sleep(2*time.Second)
		}
	}()
}

func packProc(stick *stick.Stick)  {
	defer stick.Close()
	for  {
		buff, err := stick.Read()
		if err != nil {
			log.Println(err.Error())
			return
		}

		packCnt++
		packSize+= int64(len(buff))

		err = stick.Write(buff)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}

func Server()  {
	lis, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Println(err.Error())
		return
	}

	defer lis.Close()

	for  {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		log.Printf("new connect from %s\n", conn.RemoteAddr().String())
		conn2 := stick.NewStick(conn)
		go packProc(conn2)
	}
}

func Client()  {
	conn, err := net.Dial("tcp", ADDR)
	if err != nil {
		log.Println(err.Error())
		return
	}
	conns := stick.NewStick(conn)
	conns.Write(make([]byte, 1500))
	packProc(conns)
}

func main()  {
	flag.Parse()
	if HELP {
		flag.Usage()
		return
	}

	if SERVER {
		Server()
	} else {
		Client()
	}
}
