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
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	WIDTH  = 100
	HEIGHT = 50
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("Ошибка завершения %q", err)
		}
	}()

	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("Was sent: %s", message)
		c.send <- message
	}
}

func processMessage(r Request, write func([]byte)) []byte {
	switch r.Cmd {
	case "init":
		go DrawMap(write) /*DrawMap*/
	case "entity":
		go GeneratePlants(write)          /*DrawPlant*/
		go GenerateHerbivoreAnimal(write) /*GenerateHerbivoreAnimal*/
	case "info":
		go GetInfoAbout(write, r.Id) /*InfoAbout*/
	}
	return []byte("ERR")
}

type Request struct {
	Cmd string
	Id  int
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			log.Printf("Ошибка завершения в writePump %q", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Printf("Ошибка ошибки %q", err)
				return
			}

			lock := sync.RWMutex{}
			r := Request{}
			err := json.Unmarshal(message, &r)

			if err != nil {
				log.Printf("С клиента пришла какая то ерунда, %q", err)
				return
			}

			processMessage(r, func(bytes []byte) {
				lock.Lock()
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Printf("Не удалось получить writer: %q", err)
					return
				}
				_, err = w.Write(bytes)
				if err != nil {
					log.Printf("Произошел какой то ой %q, %s не отправлены",
						err, string(bytes))
				} else {
					log.Printf("Байты успешно переданы %s", string(bytes))
				}
				if err := w.Close(); err != nil {
					return
				}
				lock.Unlock()
			})
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Произошел ой %q", err)
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}

	go client.writePump()
	go client.readPump()
}
