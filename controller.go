package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

//type InitResponse struct {
//	Gap   [][]int
//	OnCmd Command
//}

//func NewInitResponse() *InitResponse {
//	log.Printf("gap[%d][%d]", WIDTH, HEIGHT)
//	var resp = InitResponse{OnCmd: DrawMapCmd}
//	resp.Gap = make([][]int, HEIGHT)
//	for i := 0; i < HEIGHT; i++ {
//		resp.Gap[i] = make([]int, WIDTH)
//		for j := 0; j < WIDTH; j++ {
//			resp.Gap[i][j] = rand.Intn(1)
//		}
//	}
//	return &resp
//}

//func DrawMap() {
//	/* func to init gap with different colours */
//	/* todo: bonus */
//
//	var resp = NewInitResponse()
//	r, er := json.Marshal(resp)
//	if er != nil {
//		log.Printf("Возникли ошибки при маршалинге %q", er)
//		return
//	}
//	write(r)
//}

var globId = 0
var lockForId = sync.Mutex{}

func getNextId() int {
	lockForId.Lock()
	defer lockForId.Unlock()
	globId++
	return globId
}

func addPlant() {
	pl := &Plant{
		BaseEntity: BaseEntity{
			Id:   getNextId(),
			Top:  rand.Intn(AllHeight),
			Left: rand.Intn(AllWidth),
		},
		Type: rand.Intn(6),
	}
	StoragePlants[pl.Id] = pl
	data := pl.AsCmdToJs()
	if data != nil {
		write(data)
	}
}

func GeneratePlants() {
	for count := 10 + rand.Int()%15; count > 0; count-- {
		addPlant()
	}
}

var StoragePlants = make(map[int]*Plant)
var StorageHerbivoreAnimal = make(map[int]*HerbivoreAnimal)
var StoragePredatoryAnimal = make(map[int]*PredatoryAnimal)

type BaseEntity struct {
	Id   int
	Top  int
	Left int
}

type Plant struct {
	BaseEntity
	Type int
}

func (p *Plant) AsCmdToJs() []byte {
	type RespDrawPlant struct {
		OnCmd Command
		Top   int
		Left  int
		Type  int
		Id    int
	}

	data, err := json.Marshal(RespDrawPlant{
		OnCmd: DrawPlant,
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

func GenerateHerbivoreAnimal() {
	for count := 6 + rand.Int()%5; count > 0; count-- {
		an := &HerbivoreAnimal{
			BaseAnimal: BaseAnimal{
				BaseEntity: BaseEntity{
					Id:   getNextId(),
					Top:  rand.Intn(AllHeight),
					Left: rand.Intn(AllWidth),
				},
				Hunger: 0,
			},
			Target: nil,
		}
		StorageHerbivoreAnimal[an.Id] = an
		data := an.AsCmdToJs()
		if data != nil {
			write(data)
			go an.MoveInTheBackground()
			go an.StarveInTheBackground()
		}
	}
}

type BaseAnimal struct {
	BaseEntity
	Hunger int /*На сколько сильно голоден из 100. если 100 умирает*/
	Target *BaseEntity
}

type HerbivoreAnimal struct {
	BaseAnimal
	Target *Plant
}

func GetInfoAbout(id int) {
	if _, ok := StoragePlants[id]; ok {
		type ResponsePlants struct {
			OnCmd Command
			Class string
		}
		js, err := json.Marshal(ResponsePlants{
			Class: "Plant",
			OnCmd: InfoAbout,
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
			OnCmd  Command
			Class  string
			Hunger int
			Target *Plant
		}
		js, err := json.Marshal(ResponsePlants{
			Class:  "ResponsePlants",
			Hunger: data.Hunger,
			Target: data.Target,
			OnCmd:  InfoAbout,
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
		OnCmd Command
		Top   int
		Left  int
		Id    int
	}

	data, err := json.Marshal(RespDrawHerbivoreAnimal{
		OnCmd: DrawHerbivoreAnimal,
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

func (p *BaseAnimal) StarveInTheBackground() {
	ticker := time.NewTicker(StarveProcessPeriod)

	for {
		select {
		case <-LastClient.die:
			return
		case <-ticker.C:
			p.Hunger++
			if p.Hunger == 100 {
				// Не сумел найти себе еду! - умираешь
				js, err := json.Marshal(&struct {
					OnCmd  Command
					Id     int
					Reason Reason
				}{
					OnCmd:  MustDie,
					Id:     p.Id,
					Reason: Starvation,
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

func (p *BaseAnimal) MoveInTheBackground() {
	ticker := time.NewTicker(MovingPeriod)

	for {
		select {
		case <-LastClient.die:
			return
		case <-ticker.C:
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
		if data, ok := StoragePredatoryAnimal[id]; ok {
			return &HunterInfomation{
				Obj:       &data.BaseAnimal,
				Targeting: "Herbivore",
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
			c.lock.Lock()
			c.conn.WriteJSON(struct {
				OnCmd    Command
				Strategy int
				ChangeX  int
				ChangeY  int
				IdObj    int
			}{
				OnCmd:    MoveMe,
				Strategy: 0,
				ChangeX:  dirX,
				ChangeY:  dirY,
				IdObj:    id,
			})
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
			for _, data := range StoragePlants {
				if dist := func() float64 {
					return - math.Sqrt(math.Pow(float64(data.Left-o.Left), 2) +
						math.Pow(float64(data.Top-o.Top), 2))
				}(); dist < val {
					val = dist
					o.Target = &data.BaseEntity
				}
			}
			if o.Target == nil {
				log.Printf("o target всё еще nil")
				// Кушать нечего - паниковать!
				_memory[id] = struct {
					Type  int
					Value func() bool
				}{Type: -1, Value: func() bool { return false }}
				return
			}
		} else if t == "Herbivore" {
			var val float64 = 1
			for _, data := range StorageHerbivoreAnimal {
				if dist := func() float64 {
					return - math.Sqrt(math.Pow(float64(data.Left-o.Left), 2) +
						math.Pow(float64(data.Top-o.Top), 2))
				}(); dist < val {
					val = dist
					o.Target = &data.BaseEntity
				}
			}
			if o.Target == nil {
				log.Printf("o target всё еще nil")
				// Кушать нечего - паниковать!
				_memory[id] = struct {
					Type  int
					Value func() bool
				}{Type: -1, Value: func() bool { return false }}
				return
			}
		}
		var getStep = func(from int, to int) int {
			if from < to {
				return rand.Intn(5) + 2
			} else {
				return -(rand.Intn(5) + 2)
			}
		}

		_memory[id] = struct {
			Type  int
			Value func() bool
		}{Type: 1, Value: func() bool {
			if o.Target == nil {
				return false
			}
			if _, ok := StoragePlants[o.Target.Id]; !ok {
				return false
			}
			dx := getStep(o.Left, o.Target.Left)
			dy := getStep(o.Top, o.Target.Top)
			o.Left += dx
			o.Top += dy
			c.lock.Lock()
			c.conn.WriteJSON(struct {
				OnCmd    Command
				Strategy int
				ChangeX  int
				ChangeY  int
				IdObj    int
			}{
				OnCmd:    MoveMe,
				Strategy: 1,
				ChangeX:  dx,
				ChangeY:  dy,
				IdObj:    id,
			})
			c.lock.Unlock()
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
		case <-c.die:
			return
		}
	}
}

func (c *Client) KillerManager() {
	ticker := time.NewTicker(KillCheckerPeriod)

	for {
		select {
		case <-c.die:
			return
		case <-ticker.C:
			/* Травоядные животные и растения */
			for _, animal := range StorageHerbivoreAnimal {
				for id, plant := range StoragePlants {
					if meet := func(abl int, abt int, bbl int, bbt int) bool {
						in := func(a int, b int, c int) bool {
							return (a <= b) && (b <= c)
						}
						two := func(a int, b int, c int) bool {
							return in(a, b, a+c) || in(b, a, b+c)
						}
						return two(abl, bbl, EntityWidth) && two(abt, bbt, EntityHeight)
					}(animal.Left, animal.Top, plant.Left, plant.Top); meet {
						c.lock.Lock()
						c.conn.WriteJSON(struct {
							OnCmd  Command
							Id     int
							Reason Reason
						}{
							OnCmd:  MustDie,
							Id:     id,
							Reason: Eaten,
						})
						c.lock.Unlock()
						delete(StoragePlants, id)
						animal.Hunger -= 20
					}
				}
			}
		}
	}
}

type PredatoryAnimal struct {
	BaseAnimal
	Target *HerbivoreAnimal
}

func (p *PredatoryAnimal) AsCmdToJs() []byte {
	type RespDrawPredatoryAnimal struct {
		OnCmd Command
		Top   int
		Left  int
		Id    int
	}

	data, err := json.Marshal(RespDrawPredatoryAnimal{
		OnCmd: DrawPredatoryAnimal,
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

func GeneratePredatoryAnimal() {
	for count := 1 + rand.Int()%2; count > 0; count-- {
		an := &PredatoryAnimal{
			BaseAnimal: BaseAnimal{
				BaseEntity: BaseEntity{
					Id:   getNextId(),
					Top:  rand.Intn(AllHeight),
					Left: rand.Intn(AllWidth),
				},
				Hunger: 0,
			},
			Target: nil,
		}
		StoragePredatoryAnimal[an.Id] = an
		data := an.AsCmdToJs()
		if data != nil {
			write(data)
			go an.MoveInTheBackground()
			go an.StarveInTheBackground()
		}
	}
}

func (c *Client) PopulatePlants() {
	for {
		if len(StoragePlants) < 5 {
			addPlant()
		}
		select {
		case <-c.die:
			return
		}
	}
}

type Command string
type Reason string

const (
	CountX = 100
	CountY = 50

	PanelWidth  = 10
	PanelHeight = 10

	EntityWidth  = 30
	EntityHeight = 30

	AllWidth  = CountX*PanelWidth - EntityWidth
	AllHeight = CountY*PanelHeight - EntityHeight

	MovingPeriod        = 1000 * time.Millisecond
	KillCheckerPeriod   = 1000 * time.Millisecond
	StarveProcessPeriod = 1000 * time.Millisecond

	//DrawMapCmd          Command = "DrawMapCmd"
	DrawPlant           Command = "DrawPlant"
	InfoAbout           Command = "InfoAbout"
	DrawHerbivoreAnimal Command = "DrawHerbivoreAnimal"
	MoveMe              Command = "MoveMe"
	MustDie             Command = "MustDie"
	DrawPredatoryAnimal Command = "DrawPredatoryAnimal"
	Bue                 Command = "Bue"

	Starvation       Reason = "Умер от голода"
	Eaten            Reason = "Его съели"
	LimitConnections Reason = "Maximum concurrent connections exceeded"
)
