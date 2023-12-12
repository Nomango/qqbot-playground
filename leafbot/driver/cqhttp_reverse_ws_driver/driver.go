package cqhttp_reverse_ws_driver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// Driver
// @Description:
//
type Driver struct {
	Name             string
	address          string
	token            string
	bots             sync.Map
	eventChan        chan []byte
	connectHandle    func(selfId int64, host string, clientRole string)
	disConnectHandle func(selfId int64)
}

// OnConnect
/**
 * @Description:
 * @receiver d
 * @param f
 * example
 */
func (d *Driver) OnConnect(f func(selfId int64, host string, clientRole string)) {
	d.connectHandle = f
}

// OnDisConnect
/**
 * @Description:
 * @receiver d
 * @param f
 * example
 */
func (d *Driver) OnDisConnect(f func(selfId int64)) {
	d.disConnectHandle = f
}

// GetBots
/**
 * @Description:
 * @receiver d
 * @return map[int64]interface{}
 * example
 */
func (d *Driver) GetBots() map[int64]interface{} {
	m := make(map[int64]interface{})
	d.bots.Range(func(key, value interface{}) bool {
		m[key.(int64)] = value
		return true
	})

	return m
}

// ws协议
var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

// ServeHTTP
/**
 * @Description:
 * @receiver d
 * @param writer
 * @param request
 * example
 */
func (d *Driver) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	selfID, err := strconv.ParseInt(request.Header.Get("X-Self-ID"), 10, 64)
	role := request.Header.Get("X-Client-Role")
	host := request.Header.Get("Host")
	if d.token != "" {
		get := request.Header.Get("Authorization")
		auth := strings.Split(get, " ")
		if auth[0] != "Bearer" || auth[1] != d.token {
			log.Errorln("the token is not current!")
			return
		}
	}
	conn, err := upgrade.Upgrade(writer, request, nil)
	if err != nil {
		return
	}
	b := new(Bot)
	b.conn = conn
	b.selfId = selfID
	b.responses = sync.Map{}
	_, ok := d.bots.Load(selfID)
	if ok {
		d.bots.LoadOrStore(selfID, b)
	} else {
		d.bots.Store(selfID, b)
	}
	b.disConnectHandle = d.disConnectHandle
	log.Infoln(fmt.Sprintf("the bot %v is connected", selfID))
	// 执行链接回调
	go d.connectHandle(selfID, host, role)
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				b.wsClose()
				log.Errorln("ws链接读取出现错误")
				log.Errorln(err)
			}
		}()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				b.wsClose()
			}

			echo := gjson.GetBytes(data, "echo")
			if echo.Exists() {
				b.responses.Store(echo.String(), data)
			} else {
				d.eventChan <- data
			}
		}
	}()
}

// Run
/**
 * @Description:
 * @receiver d
 */
func (d *Driver) Run() {
	http.Handle("/"+d.Name+"/ws", d)
	log.Infoln("Load the cqhttp_reverse_driver successful")
	log.Infoln(fmt.Sprintf("the cqhttp_reverse_driver listening in %v", d.address))
	if err := http.ListenAndServe(d.address, nil); err != nil {
		log.Panicln(err.Error())
	}
}

// GetEvent
/**
 * @Description:
 * @receiver d
 * @return chan
 */
func (d *Driver) GetEvent() chan []byte {
	return d.eventChan
}

// GetBot
/**
 * @Description:
 * @receiver d
 * @param i
 * @return interface{}
 */
func (d *Driver) GetBot(i int64) interface{} {
	load, ok := d.bots.Load(i)
	if ok {
		return load
	}

	return nil
}

// NewDriver
/**
 * @Description:
 * @return *Driver
 */
func NewDriver(address, token string) *Driver {
	d := new(Driver)
	d.Name = "cqhttp"
	d.address = address
	d.token = token
	d.bots = sync.Map{}
	d.eventChan = make(chan []byte)
	return d
}
