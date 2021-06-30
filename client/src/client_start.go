package client

import (
	"io/ioutil"
	"log"
	"net/rpc"
	"time"

	"gopkg.in/yaml.v2"
)

type Client struct {
	index int
	conn  *rpc.Client
}

type Config struct {
	Index     int      `yaml:"index"`
	Addresses []string `yaml:"addresses"`
}

func readConfig() (*Config, error) {
	var err error

	config, err := ioutil.ReadFile("../config.yaml")
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	if err = yaml.Unmarshal(config, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getConnection(addresses []string, servers *[]*rpc.Client, finished chan bool) {
	ch := make(chan *Client)
	for i, address := range addresses {
		go func(address string, ch chan *Client, i int) {
			conn, err := rpc.DialHTTP("tcp", address)
			log.Println("try to connect", address, i)
			for err != nil {
				time.Sleep(time.Duration(time.Second))
				log.Println("try to connect again", address, i)
				conn, err = rpc.DialHTTP("tcp", address)
			}
			log.Println("connect sucessfully!", address, i)
			ch <- &Client{
				index: i,
				conn:  conn,
			}
		}(address, ch, i)
	}
	rest := len(addresses)
	for rest > 0 {
		c := <-ch
		(*servers)[c.index] = c.conn
		rest--
	}
	finished <- true
}

func Start(addresses []string) *Clerk {
	var (
		// err error
		ck *Clerk
		// in  string
	)
	servers := make([]*rpc.Client, len(addresses))
	finished := make(chan bool)
	go getConnection(addresses, &servers, finished)
	<-finished
	ck = MakeClerk(servers)
	return ck
	// reader := bufio.NewReader(os.Stdin)
	// writer := bufio.NewWriter(os.Stdout)
	// for {
	// 	writer.Write([]byte("please type again"))
	// 	in, err = reader.ReadString('\n')
	// 	if err != nil {
	// 		writer.Write([]byte("please type again"))
	// 		continue
	// 	}
	// 	args := strings.Split(in, " ")
	// 	if len(args) == 2 && args[0] == "get" {
	// 		writer.Write([]byte(ck.Get(args[1])))
	// 	} else if len(args) == 3 && args[0] == "put" {
	// 		ck.PutAppend(args[1], args[2], "put")
	// 	} else {
	// 		writer.Write([]byte("please type again"))
	// 		continue
	// 	}
	// }
}
