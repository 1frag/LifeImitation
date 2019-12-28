// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"log"
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
		log.Print(message)
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

func processMessage(r Request) {
	switch r.Cmd {
	case "init":
		//go DrawMap(write) /*DrawMap*/
	case "entity":
		go GeneratePlants()          /*DrawPlant*/
		go GenerateHerbivoreAnimal() /*GenerateHerbivoreAnimal*/
		go GeneratePredatoryAnimal() /*_EMPTY_*/
	case "info":
		go GetInfoAbout(r.Id) /*InfoAbout*/
	}
}

type Request struct {
	Cmd string
	Id  int
}

func write(bytes []byte) {
	LastClient.lock.Lock()
	w, err := LastClient.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Printf("Не удалось получить writer: %q", err)
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
	LastClient.lock.Unlock()
}

func (c *Client) writePump() {
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			r := Request{}
			err := json.Unmarshal(message, &r)

			if err != nil {
				log.Printf("С клиента пришла какая то ерунда, %q", err)
				return
			}

			processMessage(r)
		case <-c.die:
			log.Print("writePump has been closed")
			return
		default:
			continue
		}
	}
}

var LastClient *Client

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		conn: conn,
		die:  make(chan bool),
		send: make(chan []byte),
		lock: sync.RWMutex{},
	}

	if LastClient != nil {
		LastClient.lock.Lock()
		_ = LastClient.conn.WriteMessage(websocket.TextMessage, BueMessage)
		_ = LastClient.conn.Close()
		close(LastClient.die)
		StoragePlants = make(map[int]*Plant)
		StorageHerbivoreAnimal = make(map[int]*HerbivoreAnimal)
		StoragePredatoryAnimal = make(map[int]*PredatoryAnimal)
		LastClient.lock.Unlock()
	}

	LastClient = client
	go client.writePump()
	go client.readPump()
	go client.MovingManager()
	go client.KillerManager()
	go client.PopulatePlants()

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
