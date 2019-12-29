package main

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
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
			die:  make(chan bool),
		},
		Type: rand.Intn(6),
	}
	StoragePlants[pl.Id] = pl
	data := pl.AsCmdToJs()
	if data != nil {
		write(data)
	}
}

func randRange(left int, right int) int { // left <= result <= right
	return left + rand.Intn(right-left+1)
}

func GeneratePlants() {
	for i := randRange(MinCountPlants, MaxCountPlants); i > 0; i-- {
		addPlant()
	}
}

var StoragePlants MapOfPlants = make(map[int]*Plant)
var StorageHerbivoreAnimal MapOfHAnimal = make(map[int]*HerbivoreAnimal)
var StoragePredatoryAnimal MapOfPAnimal = make(map[int]*PredatoryAnimal)

type MapOfPlants map[int]*Plant
type MapOfHAnimal map[int]*HerbivoreAnimal
type MapOfPAnimal map[int]*PredatoryAnimal
type MapOfBEntity map[int]*BaseEntity

// todo: how to fix this shit?
func (s *MapOfPlants) getBaseEntity() (r MapOfBEntity) {
	r = make(map[int]*BaseEntity)
	for i, e := range *s {
		r[i] = &e.BaseEntity
	}
	return
}

func (s *MapOfHAnimal) getBaseEntity() (r MapOfBEntity) {
	r = make(map[int]*BaseEntity)
	for i, e := range *s {
		r[i] = &e.BaseEntity
	}
	return
}

func (s *MapOfPAnimal) getBaseEntity() (r MapOfBEntity) {
	r = make(map[int]*BaseEntity)
	for i, e := range *s {
		r[i] = &e.BaseEntity
	}
	return
}

func getEntity(id int) *BaseEntity {
	if o, ok := StoragePredatoryAnimal[id]; ok {
		return &o.BaseEntity
	}
	if o, ok := StorageHerbivoreAnimal[id]; ok {
		return &o.BaseEntity
	}
	if o, ok := StoragePlants[id]; ok {
		return &o.BaseEntity
	}
	return nil
}

//func StorageConvert(i interface{}) map[int]*BaseEntity {
//	listSlice, ok := i.(*map[interface{}]interface{})
//	if !ok {
//		log.Print("<>ASAKLS")
//		return nil
//	}
//	retval := make(map[int]*BaseEntity)
//	for id_, v := range *listSlice {
//		item, ok := v.(*BaseEntity)
//		if !ok {
//			log.Print("asdasdas")
//			return nil
//		}
//		if id, ok := id_.(int); ok {
//			retval[id] = item
//		}
//	}
//	return retval
//}

type BaseEntity struct {
	Id   int
	Top  int
	Left int
	die  chan bool
}

type Plant struct {
	BaseEntity
	Type int
}

func IsClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func (p *BaseEntity) remove(reason Reason) {
	if IsClosed(p.die) {
		log.Printf("Bad attempt to remove id=%d", p.Id)
		return
	}
	log.Printf("Good endlife of id=%d", p.Id)
	close(p.die)
	r := MustDieEntity{
		OnCmd:  MustDie,
		Id:     p.Id,
		Reason: reason,
	}
	LastClient.lock.Lock()
	LastClient.conn.WriteJSON(r)
	LastClient.lock.Unlock()
}

func (p *HerbivoreAnimal) remove(reason Reason) {
	if p != nil {
		p.BaseEntity.remove(reason)
		delete(StorageHerbivoreAnimal, p.Id)
	}
}

func (p *PredatoryAnimal) remove(reason Reason) {
	if p != nil {
		p.BaseEntity.remove(reason)
		delete(StoragePredatoryAnimal, p.Id)
	}
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
	for i := randRange(MinCountHAnimal, MaxCountHAnimal); i > 0; i-- {
		an := &HerbivoreAnimal{
			BaseAnimal: BaseAnimal{
				BaseEntity: BaseEntity{
					Id:   getNextId(),
					Top:  rand.Intn(AllHeight),
					Left: rand.Intn(AllWidth),
					die:  make(chan bool),
				},
				Hunger: 0,
			},
			Target: nil,
		}
		StorageHerbivoreAnimal[an.Id] = an
		data := an.AsCmdToJs()
		exist := func() bool {
			_, ok := StorageHerbivoreAnimal[an.Id]
			return ok
		}
		if data != nil {
			write(data)
			go an.MoveInTheBackground(exist)
			go an.StarveInTheBackground(exist)
		}
	}
}

type BaseAnimal struct {
	BaseEntity
	Hunger int
	Target *BaseEntity
}

type HerbivoreAnimal struct {
	BaseAnimal
	Target *Plant
}

func ServeDebug(w http.ResponseWriter, _ *http.Request) {
	/* Returns all objects in runtime now */
	d, _ := json.Marshal(struct {
		Plants  MapOfPlants
		HAnimal MapOfHAnimal
		PAnimal MapOfPAnimal
	}{
		Plants:  StoragePlants,
		HAnimal: StorageHerbivoreAnimal,
		PAnimal: StoragePredatoryAnimal,
	})
	_, _ = w.Write(d)
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
			Class:  "HerbivoreAnimal",
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
	if data, ok := StoragePredatoryAnimal[id]; ok {
		type ResponsePlants struct {
			OnCmd  Command
			Class  string
			Hunger int
			Target *HerbivoreAnimal
		}
		js, err := json.Marshal(ResponsePlants{
			Class:  "PredatoryAnimal",
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

func (p *BaseAnimal) StarveInTheBackground(exist func() bool) {
	ticker := time.NewTicker(StarveProcessPeriod)

	for exist() {
		select {
		case <-p.die:
			log.Printf("%d die", p.Id)
			return
		case <-LastClient.die:
			log.Print("StarveInTheBackground has been closed")
			return
		case <-ticker.C:
			p.Hunger++
			if p.Hunger == MaxPointLiveHunger {
				// Не сумел найти себе еду! - умираешь
				p.remove(Starvation)
			}
		}
	}
}

type MustDieEntity struct {
	OnCmd  Command
	Id     int
	Reason Reason
}

func (p *BaseAnimal) MoveInTheBackground(exist func() bool) {
	ticker := time.NewTicker(MovingPeriod)

	for exist() {
		select {
		case <-p.die:
			log.Printf("Животное %d умерло", p.Id)
			return
		case <-LastClient.die:
			log.Print("MoveInTheBackground has been closed")
			return
		case <-ticker.C:
			movingChannel <- p.Id
		}
		if !exist() {
			return
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
		if obj.Hunger >= PointToHunt {
			return 1 /* Охотится */
		}
		return 0 /* Гуляет */
	}

	var initWalk = func(id int) {
		// todo: declare to const
		var dirX = 5 - rand.Intn(11)
		var dirY = 5 - rand.Intn(11)
		var duration = 5 + rand.Intn(10)
		var smartFunc = func() bool {
			if duration == 0 {
				return false
			}
			duration--
			o := getEntity(id)
			o.Left += dirX
			o.Top += dirY
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
		if h == nil {
			return
		}
		t, o := h.Targeting, h.Obj
		var checkFunc func() bool
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
			checkFunc = func() bool {
				_, ok := StoragePlants[o.Target.Id]
				return ok
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
			checkFunc = func() bool {
				_, ok := StorageHerbivoreAnimal[o.Target.Id]
				return ok
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
			if !checkFunc() {
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
			obj_ := getHunter(id)
			if obj_ == nil {
				return
			}
			if obj := obj_.Obj; obj != nil {
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
			log.Print("MovingManager has been closed")
			return
		default:
			continue
		}
	}
}

func (c *Client) KillerManager() {
	ticker := time.NewTicker(KillCheckerPeriod)

	var areMet = func(e1l int, e1t int, e2l int, e2t int) bool {
		in := func(a int, b int, c int) bool {
			return (a <= b) && (b <= c)
		}
		two := func(a int, b int, c int) bool {
			return in(a, b, a+c) || in(b, a, b+c)
		}
		return two(e1l, e2l, EntityWidth) && two(e1t, e2t, EntityHeight)
	}

	var forr = func(power MapOfBEntity, weak MapOfBEntity,
		onMet func(int, int)) {
		if power == nil || weak == nil {
			log.Print("Bad calling `forr` function")
			return
		}
		for id1, e1 := range power {
			for id2, e2 := range weak {
				if !areMet(e1.Left, e1.Top, e2.Left, e2.Top) {
					continue
				}
				log.Printf("1l=%d 1t=%d 2l=%d 2t=%d", e1.Left, e1.Top, e2.Left, e2.Top)
				onMet(id1, id2)
			}
		}
	}

	for {
		select {
		case <-c.die:
			log.Print("KillerManager has been closed")
			return
		case <-ticker.C:
			/* Травоядные животные и растения */
			forr(StorageHerbivoreAnimal.getBaseEntity(),
				StoragePlants.getBaseEntity(), func(pid int, wid int) {
					StoragePlants[wid].remove(Eaten)
					StorageHerbivoreAnimal[pid].Hunger -= RidPointHungerIfKill
				})
			/* Хищные животные и травояденые */
			forr(StoragePredatoryAnimal.getBaseEntity(),
				StorageHerbivoreAnimal.getBaseEntity(), func(pid int, wid int) {
					log.Print("ertyuiosadkjasmd")
					StorageHerbivoreAnimal[wid].remove(Eaten)
					StoragePredatoryAnimal[pid].Hunger -= RidPointHungerIfKill
				})
		default:
			continue
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
	for i := randRange(MinCountPAnimal, MaxCountPAnimal); i > 0; i-- {
		an := &PredatoryAnimal{
			BaseAnimal: BaseAnimal{
				BaseEntity: BaseEntity{
					Id:   getNextId(),
					Top:  rand.Intn(AllHeight),
					Left: rand.Intn(AllWidth),
					die:  make(chan bool),
				},
				Hunger: 0,
			},
			Target: nil,
		}
		StoragePredatoryAnimal[an.Id] = an
		data := an.AsCmdToJs()
		exist := func() bool {
			_, ok := StoragePredatoryAnimal[an.Id]
			return ok
		}
		if data != nil {
			write(data)
			go an.MoveInTheBackground(exist)
			go an.StarveInTheBackground(exist)
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
			log.Print("PopulatePlants has been closed")
			return
		default:
			continue
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
	KillCheckerPeriod   = 500 * time.Millisecond
	StarveProcessPeriod = 500 * time.Millisecond

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

	RidPointHungerIfKill = 20
	PointToHunt          = 20
	MaxPointLiveHunger   = 100
	MinCountPlants       = 1
	MaxCountPlants       = 5
	MinCountHAnimal      = 10
	MaxCountHAnimal      = 12
	MinCountPAnimal      = 10
	MaxCountPAnimal      = 12
)
