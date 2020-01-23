// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait = 60 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
	lock sync.RWMutex
	die  chan bool
}

func (c *Client) readPump() {
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		if message == nil {
			continue
		}
		c.send <- message
		select {
		case <-c.die:
			log.Print("readPump has been closed")
			return
		default:
			continue
		}
	}
}

func ProcessMessage(r ClientMessage) {
	switch r.Cmd {
	case "init":
		//go DrawMap(write) /*DrawMap*/
	case "entity":
		go GeneratePlants()
		go GenerateAnimals()
		go GeneratePeople()
	case "info":
		GetInfoAbout(r.Id)
	}
}

type ClientMessage struct {
	Cmd string
	Id  int
}

func write(bytes []byte) {
	LastClient.lock.Lock()
	defer LastClient.lock.Unlock()
	w, err := LastClient.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Printf("Не удалось получить writer: %q", err)
		if !IsClosed(LastClient.die) {
			LastClient.die <-true
		}
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Printf("Произошел какой то ой %q, %s не отправлены",
			err, string(bytes))
	}
	if err := w.Close(); err != nil {
		return
	}
}

func WriteJSON(i interface{}) {
	d, err := json.Marshal(i)
	if err != nil {
		log.Print(err)
		return
	}
	write(d)
}

func (c *Client) writePump() {
	reason := ""
	defer log.Printf("writePump has been closed (reason=%s)", reason)
	for {
		select {
		case message := <-c.send:
			r := ClientMessage{}
			err := json.Unmarshal(message, &r)

			if err != nil {
				log.Printf("С клиента пришла какая то ерунда, %q", err)
				reason = "err != nil"
				return
			}
			log.Printf("{Id: %d, Cmd: %s}", r.Id, r.Cmd)
			ProcessMessage(r)
		case <-c.die:
			reason = "c.die"
			return
		}
	}
}

var LastClient *Client

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Print("func serveWs")
	rand.Seed(123)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "Unable to upgrade to websockets"
		log.Print(err)
		http.Error(w, m, http.StatusBadRequest)
		return
	}
	client := &Client{
		conn: conn,
		die:  make(chan bool),
		send: make(chan []byte),
		lock: sync.RWMutex{},
	}
	storage = NewStorage()

	if LastClient != nil {
		func() {
			log.Print("Last client has been killed")
			LastClient.lock.Lock()
			storage.lock.Lock()
			defer LastClient.lock.Unlock()
			defer storage.lock.Unlock()
				_ = LastClient.conn.WriteMessage(websocket.TextMessage, BueMessage)
			_ = LastClient.conn.Close()
			close(LastClient.die)
			globId = 0
		}()
	}

	log.Print("Start new game")
	LastClient = client
	go client.writePump()
	go client.readPump()
	go client.StarveInTheBackground()
	go client.MovingManager()
	go client.MeetingManager()
	go client.Populate()

}

/* Unfortunately now we support not more than one opening connection.
The latest connection will be alive. Others should receive this error message
*/
type CmdBue struct {
	OnCmd  Command
	Reason Reason
}

var BueMessage, _ = json.Marshal(CmdBue{
	OnCmd:  Bue,
	Reason: LimitConnections,
})
