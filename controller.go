package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

type InitResponse struct {
	Gap   [][]int
	OnCmd string
}

func NewInitResponse() *InitResponse {
	log.Printf("gap[%d][%d]", WIDTH, HEIGHT)
	var resp = InitResponse{OnCmd: "DrawMap"}
	resp.Gap = make([][]int, HEIGHT)
	for i := 0; i < HEIGHT; i++ {
		resp.Gap[i] = make([]int, WIDTH)
		for j := 0; j < WIDTH; j++ {
			resp.Gap[i][j] = rand.Intn(1)
		}
	}
	return &resp
}

func DrawMap(write func([]byte)) {
	/* func to init gap with different colours */
	/* todo: bonus */

	var resp = NewInitResponse()
	r, er := json.Marshal(resp)
	if er != nil {
		log.Printf("Возникли ошибки при маршалинге %q", er)
		return
	}
	write(r)
}

var globId = 0

func getNextId() int {
	globId++
	return globId
}

func GeneratePlants(write func([]byte)) {

	for count := 10 + rand.Int()%15; count > 0; count-- {
		pl := Plant{
			BaseEntity: BaseEntity{
				Id:   getNextId(),
				Top:  rand.Float64(),
				Left: rand.Float64(),
			},
			Type: rand.Intn(6),
		}
		StoragePlants[pl.Id] = pl
		data := pl.AsCmdToJs()
		if data != nil {
			write(data)
		}
	}
}

var StoragePlants = make(map[int]Plant)
var StorageHerbivoreAnimal = make(map[int]HerbivoreAnimal)

type BaseEntity struct {
	Id   int
	Top  float64
	Left float64
}

type Plant struct {
	BaseEntity
	Type int
}

func (p *Plant) AsCmdToJs() []byte {
	type RespDrawPlant struct {
		OnCmd string
		Top   float64
		Left  float64
		Type  int
		Id    int
	}

	data, err := json.Marshal(RespDrawPlant{
		OnCmd: "DrawPlant",
		Top:   p.Top,
		Left:  p.Left,
		Type:  p.Type,
		Id:    p.Id,
	})
	if err != nil {
		log.Printf("Ошибка при маршале %q", err)
		return nil
	}
	return data
}

func GenerateHerbivoreAnimal(write func([]byte)) {
	for count := 6 + rand.Int()%5; count > 0; count-- {
		an := HerbivoreAnimal{
			BaseAnimal: BaseAnimal{
				BaseEntity: BaseEntity{
					Id:   getNextId(),
					Top:  rand.Float64(),
					Left: rand.Float64(),
				},
				Hunger: 0,
			},
			Target: nil,
		}
		StorageHerbivoreAnimal[an.Id] = an
		data := an.AsCmdToJs()
		if data != nil {
			write(data)
			an.MoveInTheBackground(write)
		}
	}
}

type ParallelizationMechanism struct {
	/* Механизм распараллеливания функций для
			выполнения в задач фоне*/
	ticker *time.Ticker
	channel chan bool
}

type BaseAnimal struct {
	BaseEntity
	Hunger int /*На сколько сильно голоден из 100. если 100 умирает*/
	Target interface{}
	moving *ParallelizationMechanism
}

type HerbivoreAnimal struct {
	BaseAnimal
	Target *Plant
}

func GetInfoAbout(write func([]byte), id int) {
	if _, ok := StoragePlants[id]; ok {
		type ResponsePlants struct {
			OnCmd string
			Class string
		}
		js, err := json.Marshal(ResponsePlants{
			Class: "Plant",
			OnCmd: "InfoAbout",
		})
		if err != nil {
			log.Print(err)
			return
		}
		write(js)
		return
	}
	if data, ok := StorageHerbivoreAnimal[id]; ok {
		type ResponsePlants struct {
			OnCmd  string
			Class  string
			Hunger int
			Target *Plant
		}
		js, err := json.Marshal(ResponsePlants{
			Class:  "ResponsePlants",
			Hunger: data.Hunger,
			Target: data.Target,
			OnCmd:  "InfoAbout",
		})
		if err != nil {
			log.Print(err)
			return
		}
		write(js)
		return
	}
	log.Printf("%d не нашел нигде", id)
}

func (p *HerbivoreAnimal) AsCmdToJs() []byte {
	type RespDrawHerbivoreAnimal struct {
		OnCmd string
		Top   float64
		Left  float64
		Id    int
	}

	data, err := json.Marshal(RespDrawHerbivoreAnimal{
		OnCmd: "DrawHerbivoreAnimal",
		Top:   p.Top,
		Left:  p.Left,
		Id:    p.Id,
	})
	if err != nil {
		log.Printf("Ошибка при маршале %q", err)
		return nil
	}
	return data
}

func (p *HerbivoreAnimal) MoveInTheBackground(write func([]byte)) {
	p.moving = &ParallelizationMechanism{
		ticker:  time.NewTicker(500 * time.Millisecond),
		channel: make(chan bool),
	}

	for {
		select {
		case <-p.moving.channel:
			return
		case _ = <-p.moving.ticker.C:
			// todo: написать некоторую оболочку для движения
			//  тут только послеть коммманду что ему пора.
		}
	}
}
