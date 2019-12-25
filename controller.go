package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
var lockForId = sync.Mutex{}

func getNextId() int {
	lockForId.Lock()
	log.Print(globId)
	defer lockForId.Unlock()
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
			go an.MoveInTheBackground(write)
			go an.StarveInTheBackground(write)
		}
	}
}

type ParallelizationMechanism struct {
	/* Механизм распараллеливания функций для
	выполнения в задач фоне*/
	ticker  *time.Ticker
	channel chan bool
}

type BaseAnimal struct {
	BaseEntity
	Hunger     int /*На сколько сильно голоден из 100. если 100 умирает*/
	Target     *BaseEntity
	moving     *ParallelizationMechanism
	starvation *ParallelizationMechanism
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

func (p *HerbivoreAnimal) StarveInTheBackground(write func([]byte)) {
	p.starvation = &ParallelizationMechanism{
		ticker:  time.NewTicker(500 * time.Millisecond),
		channel: make(chan bool),
	}

	for {
		select {
		case <-p.starvation.channel:
			return
		case _ = <-p.starvation.ticker.C:
			//p.Hunger++
			if p.Hunger == 100 {
				// Не сумел найти себе еду! - умираешь
				js, err := json.Marshal(&struct {
					OnCmd  string
					Id     int
					Reason string
				}{
					OnCmd:  "MustDie",
					Id:     p.Id,
					Reason: "Умер от голода",
				})
				if err != nil {
					log.Print(err)
				} else {
					write(js)
				}
			}
		}
	}
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
			movingChannel <- p.Id
		}
	}
}

var movingChannel = make(chan int)

func (c *Client) MovingManager() {
	var _memory = make(map[int]struct {
		Type  int
		Value func() bool
	})

	type Action struct {
		Init func(id int) /*заполняет память функцией степ,
		инициализируя для неё парамметры*/
		Step func() bool /*отправляет клиенту байты,
		возвращая ответ будет ли продолжение*/
	}

	type HunterInfomation struct {
		Obj       *BaseAnimal
		Targeting string
	}

	var getHunter = func(id int) *HunterInfomation {
		if data, ok := StorageHerbivoreAnimal[id]; ok {
			return &HunterInfomation{
				Obj:       &data.BaseAnimal,
				Targeting: "Plant",
			}
		}
		log.Printf("Не могу найти по id=%d животного", id)
		return nil
	}

	var getStrategy = func(obj *BaseAnimal) int {
		if obj.Hunger >= 20 {
			return 1 /* Охотится */
		}
		return 0 /* Гуляет */
	}

	var initWalk = func(id int) {
		var dirX = 5 - rand.Intn(11)
		var dirY = 5 - rand.Intn(11)
		var duration = 5 + rand.Intn(10)
		var smartFunc = func() bool {
			if duration == 0 {
				return false
			}
			duration--
			data, err := json.Marshal(struct {
				OnCmd    string
				Strategy int
				ChangeX  int
				ChangeY  int
				IdObj    int
			}{
				OnCmd:    "MoveMe",
				Strategy: 0,
				ChangeX:  dirX,
				ChangeY:  dirY,
				IdObj:    id,
			})
			if err != nil {
				log.Print(err)
			}
			c.lock.Lock()
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return false
			}
			_, _ = w.Write(data)
			c.lock.Unlock()
			return true
		}
		_memory[id] = struct {
			Type  int
			Value func() bool
		}{Type: 0, Value: smartFunc}
	}

	var initHunt = func(id int) {
		h := getHunter(id)
		t, o := h.Targeting, h.Obj
		if t == "Plant" {
			var val float64 = 1
			var mostSuitable *int
			for id, data := range StoragePlants {
				if dist := func() float64 {
					return - math.Sqrt(math.Pow(data.Left-o.Left, 2) +
						math.Pow(data.Top-o.Top, 2))
				}(); dist < val {
					val = dist
					mostSuitable = &id
				}
			}
			var _ok_ = false
			if mostSuitable != nil {
				if data, ok := StoragePlants[*mostSuitable]; ok {
					o.Target = &data.BaseEntity
					_ok_ = ok
				}
			}
			if _ok_ {
				// Кушать нечего - паниковать!
				_memory[id] = struct {
					Type  int
					Value func() bool
				}{Type: -1, Value: func() bool { return false }}
				return
			}
		}
		var getStep = func(from float64, to float64) int {
			if from > to {
				return rand.Intn(10) + 10
			} else {
				return -(rand.Intn(10) + 10)
			}
		}

		_memory[id] = struct {
			Type  int
			Value func() bool
		}{Type: 1, Value: func() bool {
			if o.Target == nil {
				return false
			}
			data, err := json.Marshal(struct {
				OnCmd    string
				Strategy int
				ChangeX  int
				ChangeY  int
				IdObj    int
			}{
				OnCmd:    "MoveMe",
				Strategy: 1,
				ChangeX:  getStep(o.Left, o.Target.Left),
				ChangeY:  getStep(o.Left, o.Target.Left),
				IdObj:    id,
			})
			if err != nil {
				log.Print(err)
			}
			c.send <- data
			return true
		}}
	}

	for {
		select {
		case id := <-movingChannel:
			if obj := getHunter(id).Obj; obj != nil {
				strategy := getStrategy(obj)
				var doStep = false
				if data, ok := _memory[id]; ok {
					if data.Type == strategy {
						doStep = data.Value()
					}
				}
				/* Пушка */
				func(a *Action) {
					if doStep {
						a.Step()
					} else {
						a.Init(id)
					}
				}(func() *Action {
					if strategy == 0 {
						return &Action{Init: initWalk, Step: _memory[id].Value}
					} else if strategy == 1 {
						return &Action{Init: initHunt, Step: _memory[id].Value}
					}
					log.Print("Unexpected action")
					return nil
				}())
			} else {
				log.Printf("id=%d не хочет ходить", id)
			}
		}
	}
}
