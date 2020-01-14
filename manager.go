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
		go GeneratePredatoryAnimal()
		go GeneratePeoples()
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
		LastClient.die <-true
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

func writeJSON(i interface{}) {
	d, err := json.Marshal(i)
	if err != nil {
		log.Print(err)
		return
	}
	write(d)
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

	if LastClient != nil {
		log.Print("Last client has been killed")
		LastClient.lock.Lock()
		log.Print("1")
		_ = LastClient.conn.WriteMessage(websocket.TextMessage, BueMessage)
		log.Print("2")
		_ = LastClient.conn.Close()
		log.Print("3")
		close(LastClient.die)
		log.Print("4")
		globId = 0
		log.Print("5")
		StoragePlants = make(map[int]*Plant)
		StorageHerbivoreAnimal = make(map[int]*HerbivoreAnimal)
		StoragePredatoryAnimal = make(map[int]*PredatoryAnimal)
		StorageHouses = make(map[int]*House)
		StoragePeople = make(map[int]*People)
		log.Print("6")
		LastClient.lock.Unlock()
		log.Print("7")
	}

	log.Print("Start new game")
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
