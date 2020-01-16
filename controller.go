package main

import "C"
import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var globId = 0
var storage *Storage = nil
var lockForId = sync.Mutex{}

func getNextId() int {
	lockForId.Lock()
	defer lockForId.Unlock()
	globId++
	return globId
}

func GenerateBaseEntity(k TypeEntity) *_BaseEntity {
	return &_BaseEntity{
		Id:   getNextId(),
		Top:  rand.Intn(AllHeight),
		Left: rand.Intn(AllWidth),
		Kind: k,
		die:  make(chan bool, 2),
	}
}

func addPlant() {
	t := rand.Intn(3)
	writeJSON(struct {
		OnCmd Command
		Data  interface{}
	}{
		DrawPlant,
		func() interface{} {
			if t == 0 {
				o := Cabbage{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Cabbage),
					},
				}
				storage.addCabbage(&o)
				return o
			} else if t == 1 {
				o := Bush{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Bush),
					},
				}
				storage.addBush(&o)
				return o
			} else if t == 2 {
				o := Carrot{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Carrot),
					},
				}
				storage.addCarrot(&o)
				return o
			}
			log.Fatal("Рандомное число от 0 до 3 не принадлежит этому промежутку")
			return nil
		}(),
	})
}

func randRange(left int, right int) int { // left <= result <= right
	return left + rand.Intn(right-left+1)
}

func GeneratePlants() {
	for i := CountPlants; i > 0; i-- {
		addPlant()
	}
}

func (e *_BaseEntity) remove(reason Reason) {
	if IsClosed(e.die) {
		log.Printf("Bad attempt to remove id=%d", e.Id)
		return
	}
	close(e.die)
	writeJSON(MustDieEntity{
		OnCmd:  MustDie,
		Id:     e.Id,
		Reason: reason,
	})
}

type JSONForDrawEntity struct {
	OnCmd Command
	Top   int
	Left  int
	Type  int
	Id    int
}

func GenerateAnimals() {
	for i := CountRabbits; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Rabbit{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Rabbit),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addRabbit(an)
			return &an._BaseAnimal
		})
	}
	for i := CountZebras; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Zebra{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Zebra),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addZebra(an)
			return &an._BaseAnimal
		})
	}
	for i := CountWolfs; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Wolf{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Wolf),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addWolf(an)
			return &an._BaseAnimal
		})
	}
	for i := CountBears; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Bear{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Bear),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addBear(an)
			return &an._BaseAnimal
		})
	}
	for i := CountFoxes; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Fox{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Fox),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addFox(an)
			return &an._BaseAnimal
		})
	}
	for i := CountElephants; i > 0; i-- {
		addAnimal(func() *_BaseAnimal {
			an := &Elephant{
				_BaseAnimal: _BaseAnimal{
					_BaseEntity: *GenerateBaseEntity(_Elephant),
					Hunger:      0,
					Target:      nil,
				},
			}
			storage.addElephant(an)
			return &an._BaseAnimal
		})
	}
}

func ServeDebug(w http.ResponseWriter, _ *http.Request) {
	/* Returns all objects in runtime now */
	var data = make(map[int]map[string]string)
	for i, o := range storage.AllPlants() {
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		about["type"] = o.Kind.String()

		data[i] = about
	}
	for i, o := range storage.AllAnimal() {
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		about["hunger"] = strconv.Itoa(o.Hunger)
		data[i] = about
	}
	for i, o := range storage.AllPeople() {
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		about["hunger"] = strconv.Itoa(o.Hunger)
		if o.Target == nil {
			about["target"] = "__nil__"
		} else {
			about["target"] = strconv.Itoa(o.Target.Id)
		}
		data[i] = about
	}
	for i, o := range storage.AllHouses() {
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		if o.Husband == nil {
			about["Husband"] = "__nil__"
		} else {
			about["Husband"] = strconv.Itoa(o.Husband.Id)
		}
		if o.Wife == nil {
			about["Wife"] = "__nil__"
		} else {
			about["Wife"] = strconv.Itoa(o.Wife.Id)
		}
		data[i] = about
	}
	d, _ := json.Marshal(data)
	_, _ = w.Write(d)
}

func GetInfoAbout(id int) {
	if data := storage.GetPlantById(id); data != nil {
		writeJSON(struct {
			OnCmd Command
			Class string
		}{
			Class: data.Kind.String(),
			OnCmd: InfoAbout,
		})
		return
	}
	if data := storage.GetAnimalById(id); data != nil {
		writeJSON(struct {
			OnCmd  Command
			Class  string
			Hunger int
			Target *int
		}{
			Class:  data.Kind.String(),
			Hunger: data.Hunger,
			Target: func() *int {
				if data.Target == nil {
					return nil
				}
				return &data.Target.Id
			}(),
			OnCmd: InfoAbout,
		})
		return
	}
	if data := storage.GetHumanById(id); data != nil {
		writeJSON(struct {
			OnCmd  Command
			Class  string
			Hunger int
			Target *int
			Age    int
		}{
			OnCmd:  InfoAbout,
			Class:  "Human",
			Hunger: data.Hunger,
			Target: func() *int {
				if data.Target == nil {
					return nil
				}
				return &data.Target.Id
			}(),
			Age: data.Age,
		})
		return
	}
	if data := storage.GetHouseById(id); data != nil {
		writeJSON(struct {
			OnCmd   Command
			Class   string
			Wife    *int
			Husband *int
		}{
			OnCmd: InfoAbout,
			Class: "House",
			Wife: func() *int {
				if data.Wife == nil {
					return nil
				}
				return &data.Wife.Id
			}(),
			Husband: func() *int {
				if data.Husband == nil {
					return nil
				}
				return &data.Husband.Id
			}(),
		})
		return
	}
	log.Printf("%d не нашел нигде", id)
}

func addAnimal(creating func() *_BaseAnimal) {
	an := creating()

	writeJSON(struct {
		OnCmd Command
		Class string
		Top   int
		Left  int
		Id    int
	}{
		OnCmd: DrawAnimal,
		Class: an.Kind.String(),
		Top:   an.Top,
		Left:  an.Left,
		Id:    an.Id,
	})

	exist := func() bool {
		return storage.ExistId(an.Id)
	}

	if exist() {
		go an.MoveInTheBackground(exist)
		ChanStarve <- func() {
			an.Hunger++
		}
	}
}

var ChanStarve = make(chan func(), EntitiesLimit)

func (c *Client) StarveInTheBackground() {
	ticker := time.NewTicker(StarveProcessPeriod)
	var callbacks = make([]func(), 0)

	for !IsClosed(LastClient.die) {
		select {
		case fn := <-ChanStarve:
			callbacks = append(callbacks, fn)
		case <-ticker.C:
			for _, fn := range callbacks {
				fn()
			}
		}
	}
	log.Print("StarveInTheBackground has been closed")
}

type MustDieEntity struct {
	OnCmd  Command
	Id     int
	Reason Reason
}

const Male Gender = "Male"
const Female Gender = "Female"

func addPeople(g Gender, a int) {
	p := &Human{
		_BaseAnimal: _BaseAnimal{
			_BaseEntity: *GenerateBaseEntity(_Human),
			Hunger:      0,
			Target:      nil,
		},
		Age:      a,
		State:    0,
		Gender:   g,
		Telegram: make(chan TelegramMessage, 20),
	}
	writeJSON(p.asCmdCreate())
	storage.addHuman(p)
	go p.LifeCycle()
}

func (p *Human) LifeCycle() {
	goToObject := func(target *_BaseEntity, hunger float64) {
		l, t := target.Left, target.Top
		var dl = approach(p.Left, l, 5)
		var dt = approach(p.Top, t, 5)
		p.Left += dl
		p.Top += dt
		p.SendMoveMe(dl, dt, hunger, p.Age)
	}

	ticker := time.NewTicker(LifeCyclePeriod)
	p.SocialStatus = Child
	var home *House = nil
	var partnerAtHome = false
	var foodForHuman = []TypeEntity{
		_Rabbit, _Wolf, _Human, _House, _Bear, _Zebra,
		_Fox, _Cabbage, _Bush, _Carrot, _Elephant,
	}

	for !IsClosed(p.die) {
		select {
		case <-p.die:
			log.Printf("Животное %d умерло", p.Id)
			return
		case <-LastClient.die:
			log.Print("LifeCycle has been closed")
			return
		case <-ticker.C:
			if p.State < 6 {
				select {
				case m := <-p.Telegram:
					if m.Head == KillFood {
						p.State += 4
					} else {
						log.Fatal("Unexpected Head Message")
					}
					if p.State > 6 {
						p.State = 7
					}
				default:
					// найди ближайшего и иди к нему
					goToObject(p.nearest(foodForHuman), (7.0-float64(p.State))/8.0)
				}
			} else if p.State == 7 {
				log.Printf("%d has p.State==7 G=%s S=%s", p.Id, p.Gender, p.SocialStatus)
				// ищет себе вторую половинку
				onMet := func(o *Human) {
					goToObject(&o._BaseEntity, 0)
					p.SocialStatus = InTheWay
					p.Target = o
					p.State = 8
				}

				p.SocialStatus = InSearch
				for _, o := range storage.AllPeople() {
					if o.Gender != p.Gender && o.SocialStatus == InSearch {
						o.Telegram <- p.MessageWithSign(LetSGetMarried, nil)
						onMet(o)
						break
					}
				}
				if p.Target == nil {
					goToObject(p.nearest(foodForHuman), 0)
				}
				select {
				case m := <-p.Telegram:
					if m.Head == LetSGetMarried {
						if pTarget := storage.GetHumanById(m.From); pTarget != nil {
							onMet(pTarget)
							continue
						}
					}
				default:
					continue
				}
			} else if p.State == 8 {
				if p.Target == nil {
					p.State = 7
					continue
				}
				if !storage.ExistId(p.Target.Id) {
					p.State = 7
					continue
				}
				dl := absForInt(p.Left - p.Target.Left)
				dt := absForInt(p.Top - p.Target.Top)
				if 0 <= dl && dl <= EntityWidth && 0 <= dt && dt <= EntityHeight {
					p.State = 9
					p.Target.Telegram <- p.MessageWithSign(GoToState9, nil)
					continue
				} else {
					goToObject(&p.Target._BaseEntity, 0)
				}
				select {
				case m := <-p.Telegram:
					if m.Head == GoToState9 {
						p.State = 9
						continue
					}
				default:
					continue
				}
			} else if p.State == 9 {
				p.SocialStatus = InMarriage
				if p.Gender == Male {
					if home != nil {
						if dist(&p._BaseEntity, &home._BaseEntity) < 50 {
							// дом построен
							writeJSON(JSONForDrawEntity{
								OnCmd: DrawHouse,
								Top:   home.Top,
								Left:  home.Left,
								Id:    home.Id,
							})
							log.Printf("Built house %d (%d + %d)", home.Id,
								p.Id, p.Target.Id)
							storage.addHouse(home)
							p.Target.Telegram <- p.MessageWithSign(HouseHasBuilt, home.Id)
							p.State = 10
							p.House = home
						} else {
							// далеко, -> подойди поближе к дому чтобы его построить
							goToObject(&home._BaseEntity, 0)
						}
					} else {
						// А где строить то?
						// todo: 400, 50, 30
						h := p.nearest([]TypeEntity{_House})
						var newHouse = House{
							_BaseEntity: *GenerateBaseEntity(_House),
							Wife:        p.Target,
							Husband:     p,
						}
						nearby := &p._BaseEntity
						if h != nil && dist(h, &p._BaseEntity) < 400 {
							nearby = h
						}
						newHouse.Left = nearby.Left - 30
						newHouse.Top = nearby.Top - 50
						home = &newHouse
					}
				} else { // женщина ждет пока построют дом
					select {
					case m := <-p.Telegram:
						if m.Head == HouseHasBuilt {
							var (
								ok = false
								i  = 0
							)
							if i, ok = m.Body.(int); ok {
								p.House = storage.GetHouseById(i)
								ok = p.House != nil
							}
							if !ok {
								log.Fatal("Это никогда не должно произойти")
							}
							p.State = 10
						}
					default:
						goToObject(&_BaseEntity{
							Top:  p.Target.Left,
							Left: p.Target.Top - EntityWidth,
						}, 0)
					}
				}
			} else if p.State == 10 {
				select {
				case m := <-p.Telegram:
					if m.Head == KillFood {
						p.State = 11
						p.Target.Telegram <- p.MessageWithSign(ImGoingAtHome, nil)
					} else if m.Head == ImGoingAtHome || m.Head == IAmAtHome {
						partnerAtHome = true
					} else {
						log.Printf("Unexpected this head=(%s), target.id=%d,"+
							" from.id=%d, pah=%t", m.Head, p.Target.Id,
							m.From, partnerAtHome)
					}
				default:
					goToObject(p.nearest(foodForHuman), 0.3)
				}
			} else if p.State == 11 {
				if dist(&p._BaseEntity, p.House.Locate(p.Gender)) < 5 {
					p.State = 12
					p.Target.Telegram <- p.MessageWithSign(IAmAtHome, nil)
				} else {
					goToObject(p.House.Locate(p.Gender), 0.1)
				}
			} else if p.State == 12 {
				select {
				case m := <-p.Telegram:
					if m.Head == IAmAtHome {
						p.State = 13
					}
				}
			} else if p.State == 13 {
				if p.Gender == Female {
					p.House.CreateChild()
				}
				// конец второй части, todo: next
				goto TheEnd
			}
		}
	}
TheEnd:
	log.Printf("The end for person with id=%d", p.Id)
}

func (h *House) Locate(gender Gender) *_BaseEntity {
	if gender == Male {
		return &_BaseEntity{
			Top:  h.Left + 50,
			Left: h.Top + 30,
		}
	} else {
		return &_BaseEntity{
			Top:  h.Left + 20,
			Left: h.Top + 30,
		}
	}
}

func RandomGender() Gender {
	if rand.Intn(2) == 1 {
		return Male
	} else {
		return Female
	}
}

func (h *House) CreateChild() {
	c := &Human{
		_BaseAnimal: _BaseAnimal{
			_BaseEntity: _BaseEntity{
				Id:   getNextId(),
				Top:  h.Left + 50,
				Left: h.Top + 50,
				die:  nil,
			},
			Hunger: 50,
		},
		Age:          2, // любое число меньше 5 в js ничем не отличается
		State:        0,
		Target:       nil,
		Gender:       RandomGender(),
		SocialStatus: Child,
		Telegram:     make(chan TelegramMessage, 20),
		House:        h,
	}
	writeJSON(c.asCmdCreate())
	time.AfterFunc(20*time.Second, func() {
		c.SendMoveMe(-15, 0, 0.8, 6)
	})
	time.AfterFunc(40*time.Second, func() {
		c.SendMoveMe(-15, 0, 0.8, 15)
	})
}

func dist2(a *_BaseEntity, b *_BaseEntity) float64 {
	return math.Pow(float64(a.Left-b.Left), 2) +
		math.Pow(float64(a.Top-b.Top), 2)
}

func dist(a *_BaseEntity, b *_BaseEntity) float64 {
	return math.Pow(dist2(a, b), 0.5)
}

func (p *Human) MessageWithSign(header HeadTelegramMessage, body interface{}) TelegramMessage {
	return TelegramMessage{
		Head: header,
		From: p.Id,
		Body: body,
	}
}

func approach(from int, to int, rightSide int) int {
	if fv := func() int {
		if from < to {
			return rightSide
		} else {
			return -rightSide
		}
	}(); absForInt(fv) < absForInt(from-to) {
		return fv
	} else {
		return to - from
	}
}

func absForInt(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func (p *Human) Take(m TelegramMessage) {
	select {
	case p.Telegram <- m:
	default:
	}
}

/* My RoadMap
1. ускорить шаг любого, оргунизовав единый центр шага
*/

func getStep(id *int, kind *TypeEntity) int {
	var m = map[TypeEntity]int{
		_Rabbit:   12,
		_Wolf:     8,
		_Human:    20,
		_House:    0,
		_Bear:     4,
		_Zebra:    30,
		_Fox:      10,
		_Cabbage:  0,
		_Bush:     0,
		_Carrot:   0,
		_Elephant: 6,
	}
	if id != nil {
		if skind, ok := storage.ToType[*id]; !ok {
			log.Fatalf("В программе баг")
		} else {
			kind = &skind
		}
	}
	return m[*kind] * (1 - 2*rand.Intn(2))
}

func (e *_BaseEntity) nearest(params []TypeEntity) *_BaseEntity {
	/*
		1 - трава
		2 - зайцы
		4 - волки
		8 - люди
		16 - дома
	*/
	inner := func(s []*_BaseEntity) *_BaseEntity {
		bi, n := -1, -1
		var best *_BaseEntity = nil
		for i, o := range s {
			if o == nil {
				continue
			}
			p := (o.Left-e.Left)*(o.Left-e.Left) +
				(o.Top-e.Top)*(o.Top-e.Top)
			if bi == -1 || n > p {
				bi = i
				n = p
				best = o
			}
		}
		return best
	}
	var param = map[TypeEntity]bool{}
	for _, te := range params {
		param[te] = true
	}
	var m = make([]*_BaseEntity, 0)
	for _, d := range storage.AllBaseEntities() {
		if param[d.Kind] {
			m = append(m, d)
		}
	}
	return inner(m)
}

func (p *Human) asCmdCreate() interface{} {
	return MsgWithDrawPeople{
		OnCmd:  DrawPeople,
		Id:     p.Id,
		Top:    p.Top,
		Left:   p.Left,
		Age:    p.Age,
		Gender: p.Gender,
	}
}

type MsgWithDrawPeople struct {
	OnCmd  Command
	Id     int
	Top    int
	Left   int
	Age    int
	Gender Gender
}

func GeneratePeople() {
	for count := randRange(5, 10); count > 0; count-- {
		curGender := Male
		if rand.Intn(2) == 1 {
			curGender = Female
		}
		curAge := randRange(10, 40)
		addPeople(curGender, curAge)
	}
}

func (p *_BaseAnimal) MoveInTheBackground(exist func() bool) {
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

var movingChannel = make(chan int, 200)

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
		Obj       *_BaseAnimal
		Targeting []TypeEntity
	}

	var SafelyStep = func(o *_BaseAnimal, dx int, dy int) (nl int, nt int) {
		var c = func(left int, x int, right int) int {
			if left > x {
				return left
			}
			if right < x {
				return right
			}
			return x
		}
		nl = c(0, o.Left+dx, AllWidth-EntityWidth) - o.Left
		nt = c(0, o.Top+dy, AllHeight-EntityHeight) - o.Top
		o.Left += nl
		o.Top += nt
		return nl, nt
	}

	var getHunter = func(id int) *HunterInfomation {
		var an *_BaseAnimal
		var m = map[TypeEntity][]TypeEntity{
			_Zebra:    {_Bush},
			_Rabbit:   {_Carrot, _Cabbage},
			_Elephant: {_Bush},
			_Fox:      {_Rabbit},
			_Wolf:     {_Rabbit, _Fox},
			_Bear:     {_Rabbit, _Fox, _Wolf, _Zebra},
		}
		if an = storage.GetAnimalById(id); an != nil {
			return &HunterInfomation{
				Obj:       an,
				Targeting: m[an.Kind],
			}
		}
		return nil
	}

	var getStrategy = func(obj *_BaseAnimal) int {
		if obj.Hunger >= PointToHunt {
			return 1 /* Охотится */
		}
		return 0 /* Гуляет */
	}

	var initWalk = func(id int) {
		// todo: declare to const
		var dirX = getStep(&id, nil)
		var dirY = getStep(&id, nil)
		var duration = 5 + rand.Intn(10)
		var smartFunc = func() bool {
			if duration == 0 {
				return false
			}
			duration--
			o := storage.GetAnimalById(id)
			if o == nil {
				log.Print("storage.GetAnimalById(id) == nil !ERROR!")
				return false
			}

			dirX, dirY = SafelyStep(o, dirX, dirY)
			SendMoveMe(dirX, dirY, id, float64(o.Hunger)/float64(MaxPointLiveHunger))
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
		food := o.nearest(t)
		if food == nil {
			log.Printf("o target всё еще nil")
			// Кушать нечего - паниковать!
			_memory[id] = struct {
				Type  int
				Value func() bool
			}{Type: -1, Value: func() bool { return false }}
			return
		}
		checkFunc = func() bool {
			return storage.ExistId(food.Id)
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
			dx := approach(o.Left, o.Target.Left, absForInt(getStep(&id, nil)))
			dy := approach(o.Left, o.Target.Left, absForInt(getStep(&id, nil)))
			dx, dy = SafelyStep(o, dx, dy)
			SendMoveMe(dx, dy, id, float64(o.Hunger)/float64(MaxPointLiveHunger))
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

func (p *Human) SendMoveMe(dirX int, dirY int, hunger float64, age int) {
	writeJSON(struct {
		OnCmd   Command
		ChangeX int
		ChangeY int
		IdObj   int
		Hunger  float64
		Age     int
	}{
		OnCmd:   MoveMe,
		ChangeX: dirX,
		ChangeY: dirY,
		IdObj:   p.Id,
		Hunger:  hunger,
		Age:     age,
	})
}

func SendMoveMe(dirX int, dirY int, id int, hunger float64) {
	writeJSON(struct {
		OnCmd   Command
		ChangeX int
		ChangeY int
		IdObj   int
		Hunger  float64
	}{
		OnCmd:   MoveMe,
		ChangeX: dirX,
		ChangeY: dirY,
		IdObj:   id,
		Hunger:  hunger,
	})
}

func (c *Client) MeetingManager() {
	ticker := time.NewTicker(MeetingCheckerPeriod)

	var areMet = func(e1l int, e1t int, e2l int, e2t int) bool {
		in := func(a int, b int, c int) bool {
			return (a <= b) && (b <= c)
		}
		two := func(a int, b int, c int) bool {
			return in(a, b, a+c) || in(b, a, b+c)
		}
		return two(e1l, e2l, EntityWidth) && two(e1t, e2t, EntityHeight)
	}

	for {
		select {
		case <-c.die:
			log.Print("MeetingManager has been closed")
			return
		case <-ticker.C:
			for i, obj1 := range storage.AllBaseEntities() {
				for j, obj2 := range storage.AllBaseEntities() {
					if i <= j || !areMet(obj1.Left, obj1.Top, obj2.Left, obj2.Top) {
						continue
					}

					if (obj1.Kind == _Rabbit && obj2.Kind.in(_Carrot, _Cabbage)) ||
						(obj1.Kind == _Zebra && obj2.Kind.in(_Bush)) ||
						(obj1.Kind == _Elephant && obj2.Kind.in(_Carrot, _Cabbage, _Bush)) ||
						(obj1.Kind == _Fox && obj2.Kind.in(_Rabbit)) ||
						(obj1.Kind == _Wolf && obj2.Kind.in(_Rabbit, _Fox)) ||
						(obj1.Kind == _Bear && obj2.Kind.in(_Rabbit, _Fox, _Wolf, _Zebra)) {

						obj2.remove(Eaten)
						an := storage.GetAnimalById(obj1.Id)
						an.Hunger -= RidPointHungerIfKill
						if an.Hunger < 0 {
							an.Hunger = 0
						}
					}

					if obj1.Kind == _Human && obj2.Kind.in(_Carrot, _Cabbage,
						_Rabbit, _Wolf) {

						obj2.remove(Eaten)
						h := storage.GetHumanById(obj1.Id)
						h.Take(TelegramMessage{Head: KillFood})
					}
				}
			}
		}
	}
}

func (c *Client) Populate() {
	for ; !IsClosed(c.die); {
		if len(storage.AllPlants()) < 5 {
			addPlant()
		}
		if len(storage.AllAnimal()) < 5 {
			// todo: generate animals
		}
	}
	log.Print("Populate has been closed")
}

type Command string
type Reason string
type HeadTelegramMessage string
type TelegramMessage struct {
	Head HeadTelegramMessage
	From int
	Body interface{}
}
type SocialStatus string

const (
	CountX = 100
	CountY = 50

	EntitiesLimit  = 1000
	CountElephants = 5
	CountFoxes     = 5
	CountBears     = 5
	CountWolfs     = 5
	CountZebras    = 5
	CountRabbits   = 5

	PanelWidth  = 10
	PanelHeight = 10

	EntityWidth  = 30
	EntityHeight = 30

	RidPointHungerIfKill = 60

	AllWidth  = CountX*PanelWidth - EntityWidth
	AllHeight = CountY*PanelHeight - EntityHeight

	MovingPeriod         = 1000 * time.Millisecond
	MeetingCheckerPeriod = 500 * time.Millisecond
	StarveProcessPeriod  = 1500 * time.Millisecond
	LifeCyclePeriod      = 1000 * time.Millisecond

	Child      SocialStatus = "Ребенок еще (рано пока)"
	InSearch   SocialStatus = "В активном поиске"
	InTheWay   SocialStatus = "По пути к своей половинке"
	InMarriage SocialStatus = "Женат / замужем"

	DrawPeople          Command = "DrawPeople"
	DrawPlant           Command = "DrawPlant"
	InfoAbout           Command = "InfoAbout"
	DrawAnimal          Command = "DrawAnimal"
	MoveMe              Command = "MoveMe"
	MustDie             Command = "MustDie"
	Bue                 Command = "Bue"
	DrawHouse           Command = "DrawHouse"

	KillFood       HeadTelegramMessage = "Покушал"
	LetSGetMarried HeadTelegramMessage = "letSGetMarried"
	GoToState9     HeadTelegramMessage = "GoToState9"
	HouseHasBuilt  HeadTelegramMessage = "HouseHasBuilt"
	ImGoingAtHome  HeadTelegramMessage = "Покушал, иду домой"
	IAmAtHome      HeadTelegramMessage = "Я дома"

	Starvation       Reason = "Умер от голода"
	Eaten            Reason = "Его съели"
	LimitConnections Reason = "Maximum concurrent connections exceeded"

	PointToHunt        = 20
	MaxPointLiveHunger = 100
	CountPlants        = 15
	CountRabbit        = 12
	CountPAnimal       = 12
)
