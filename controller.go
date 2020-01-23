package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var globId = 0
var storage *Storage = nil
var lockForId = sync.Mutex{}
var helper = Helper{}

func getNextId() int {
	lockForId.Lock()
	defer lockForId.Unlock()
	globId++
	return globId
}

func GenerateBaseEntity(k TypeEntity, positions ...int) *_BaseEntity {
	retval := &_BaseEntity{
		Id:   getNextId(),
		Left: rand.Intn(AllWidth),
		Top:  rand.Intn(AllHeight),
		Kind: k,
		die:  make(chan bool, 2),
	}
	log.Printf("Created %s", retval.String())
	if len(positions) == 2 {
		retval.Left = positions[0]
		retval.Top = positions[1]
	}
	return retval
}

func addPlant() {
	WriteJSON(struct {
		OnCmd Command
		Data  interface{}
	}{
		DrawPlant,
		func() interface{} {
			if t := rand.Intn(3); t == 0 {
				o := Cabbage{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Cabbage),
					},
				}
				storage.AddCabbage(&o)
				return o
			} else if t == 1 {
				o := Bush{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Bush),
					},
				}
				storage.AddBush(&o)
				return o
			} else if t == 2 {
				o := Carrot{
					_BasePlant: _BasePlant{
						_BaseEntity: *GenerateBaseEntity(_Carrot),
					},
				}
				storage.AddCarrot(&o)
				return o
			} else {
				log.Fatalf("Fail: 0 <= %d < 3", t)
				return nil
			}
		}(),
	})
}

func GeneratePlants() {
	for i := CountPlants; i > 0; i-- {
		addPlant()
	}
}

func (e *_BaseEntity) remove(reason Reason, blocking ...bool) {
	if e.die == nil || IsClosed(e.die) {
		log.Printf("Bad attempt to remove id=%d", e.Id)
		return
	}
	close(e.die)
	if len(blocking) > 0 && !blocking[0] {
		storage.RemoveById(e.Id, false)
	} else {
		storage.RemoveById(e.Id)
	}

	WriteJSON(MustDieEntity{
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
		helper.AddRabbit = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Rabbit{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddRabbit(an)
				return &an._BaseAnimal
			})
		}
		helper.AddRabbit(GenerateBaseEntity(_Rabbit))
	}
	for i := CountZebras; i > 0; i-- {
		helper.AddZebra = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Zebra{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddZebra(an)
				return &an._BaseAnimal
			})
		}
		helper.AddZebra(GenerateBaseEntity(_Zebra))
	}
	for i := CountWolfs; i > 0; i-- {
		helper.AddWolf = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Wolf{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddWolf(an)
				return &an._BaseAnimal
			})
		}
		helper.AddWolf(GenerateBaseEntity(_Wolf))
	}
	for i := CountBears; i > 0; i-- {
		helper.AddBear = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Bear{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddBear(an)
				return &an._BaseAnimal
			})
		}
		helper.AddBear(GenerateBaseEntity(_Bear))
	}
	for i := CountFoxes; i > 0; i-- {
		helper.AddFox = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Fox{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddFox(an)
				return &an._BaseAnimal
			})
		}
		helper.AddFox(GenerateBaseEntity(_Fox))
	}
	for i := CountElephants; i > 0; i-- {
		helper.AddElephant = func(entity *_BaseEntity) int {
			return addAnimal(func() *_BaseAnimal {
				an := &Elephant{
					_BaseAnimal: _BaseAnimal{
						_BaseEntity: *entity,
						Hunger:      0,
						Target:      nil,
					},
				}
				storage.AddElephant(an)
				return &an._BaseAnimal
			})
		}
		helper.AddElephant(GenerateBaseEntity(_Elephant))
	}
	helper.AdderAnimalInitiate = true
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
		about["type"] = o.Kind.String()
		data[i] = about
	}
	for i, o := range storage.AllPeople() {
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		about["hunger"] = strconv.Itoa(o.Hunger)
		about["type"] = o.Kind.String()
		about["state"] = strconv.Itoa(o.State)
		if o.Target == nil {
			about["target"] = "__nil__"
		} else {
			about["target"] = strconv.Itoa(o.Target.Id)
		}
		if o.House == nil {
			about["home"] = "__nil__"
		} else {
			about["home"] = strconv.Itoa(o.House.Id)
		}
		data[i] = about
	}
	for i, o := range storage.AllHouses() {
		str := func(b *_BaseEntity) string {
			return fmt.Sprintf("(%d, %d)", b.Left, b.Top)
		}
		var about = make(map[string]string)
		about["left"] = strconv.Itoa(o.Left)
		about["top"] = strconv.Itoa(o.Top)
		about["type"] = o.Kind.String()
		about["Locate for Male"] = str(o.Locate(Male))
		about["Locate for Female"] = str(o.Locate(Female))
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
		WriteJSON(struct {
			OnCmd Command
			Class string
		}{
			Class: data.Kind.String(),
			OnCmd: InfoAbout,
		})
		return
	}
	if cb, data := storage.GetAnimalById(id); data != nil {
		defer cb()
		WriteJSON(struct {
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
		WriteJSON(struct {
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
		WriteJSON(struct {
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

func addAnimal(creating func() *_BaseAnimal) int {
	an := creating()

	WriteJSON(struct {
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

	MovingChannel <- an.Id
	ChanStarve <- an.Id
	return an.Id
}

var ChanStarve = make(chan int, EntitiesLimit)

func (c *Client) StarveInTheBackground() {
	ticker := time.NewTicker(StarveProcessPeriod)
	var ids = map[int]bool{}

	for !IsClosed(LastClient.die) {
		select {
		case id := <-ChanStarve:
			ids[id] = true
		case <-ticker.C:
			for id := range ids {
				cb, an := storage.GetAnimalById(id)
				if an == nil {
					delete(ids, id)
					continue
				}
				if an.Hunger == MaxPointLiveHunger {
					log.Printf("%s %s", an.String(), Starvation)
					an.remove(Starvation, false)
				} else {
					an.Hunger++
				}
				cb()
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
	WriteJSON(p.asCmdCreate())
	storage.AddHuman(p)
	go p.LifeCycle()
}

func (p *_BaseAnimal) GoToObject(target *_BaseEntity, hunger float64) {
	/* Двигает модель и отправляет результат */
	l, t := target.Left, target.Top
	var dl = approach(p.Left, l, absForInt(getStep(nil, &p.Kind)))
	var dt = approach(p.Top, t, absForInt(getStep(nil, &p.Kind)))
	p.Left += dl
	p.Top += dt
	p.SendMoveMe(dl, dt, hunger)
}

func (p *Human) LifeCycle() {
	ticker := time.NewTicker(LifeCyclePeriod)
	p.SocialStatus = Child
	var home *House = nil
	var partnerAtHome = false

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
					if food := p.NearestLst(p.Kind.GetTarget(), false); food != nil {
						p.GoToObject(food, (7.0-float64(p.State))/8.0)
					} else {
						log.Print("Еды нет в этом мире!!")
						continue
					}
				}
			} else if p.State == 7 {
				// ищет себе вторую половинку
				onMet := func(o *Human) {
					p.GoToObject(&o._BaseEntity, 0)
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
					if t := p.NearestLst(p.Kind.GetTarget(), false); t != nil {
						p.GoToObject(t, 0)
					}
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
					p.GoToObject(&p.Target._BaseEntity, 0)
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
							WriteJSON(JSONForDrawEntity{
								OnCmd: DrawHouse,
								Top:   home.Top,
								Left:  home.Left,
								Id:    home.Id,
							})
							log.Printf("Built house %d (%d + %d)", home.Id,
								p.Id, p.Target.Id)
							storage.AddHouse(home)
							p.Target.Telegram <- p.MessageWithSign(HouseHasBuilt, home.Id)
							p.State = 10
							p.House = home
						} else {
							// далеко, -> подойди поближе к дому чтобы его построить
							p.GoToObject(&home._BaseEntity, 0)
						}
					} else {
						// Решает где строить
						var newHouse = House{
							_BaseEntity: *GenerateBaseEntity(_House),
							Wife:        p.Target,
							Husband:     p,
						}
						l, t := CityBuilder(&p._BaseEntity, &p.Target._BaseEntity)
						newHouse.Left = l
						newHouse.Top = t
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
						p.GoToObject(&_BaseEntity{
							Top:  p.Target.Top,
							Left: p.Target.Left - EntityWidth,
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
					if f := p.NearestLst(p.Kind.GetTarget(), false); f != nil {
						p.GoToObject(f, 0.3)
					}
				}
			} else if p.State == 11 {
				if dist(&p._BaseEntity, p.House.Locate(p.Gender)) < 5 {
					p.State = 12
					p.Target.Telegram <- p.MessageWithSign(IAmAtHome, nil)
				} else {
					p.GoToObject(p.House.Locate(p.Gender), 0.1)
				}
			} else if p.State == 12 {
				if partnerAtHome {
					p.State = 13
					p.Target.Telegram <- p.MessageWithSign(IAmAtHome, nil)
				}
				select {
				case m := <-p.Telegram:
					if m.Head == IAmAtHome {
						p.State = 13
					}
				}
			} else if p.State == 13 {
				log.Printf("%s on state=13", p.String())
				if p.Gender == Female {
					// рожает ребенка ждет когда построют ферму
					p.House.CreateChild()
					select {
					case <-p.Telegram:
						p.State = 14
					}
					return
				} else {
					// строить забор фермы
					l, t := CityBuilder(&p.House._BaseEntity)
					p.Farm = p.BuildFarm(l, t, l+HouseWidth, t+HouseHeight)
					return
				}
			} else if p.State == 14 {
				// мужчина за скотом / женщина за овощами
				if p.Gender == Male {
					p.StompZebra()
				} else {

				}
			}
		}
	}

	log.Printf("The end for person with id=%d", p.Id)
}

func (p *Human) StompZebra() {
	timer := time.NewTicker(LifeCyclePeriod)
	z := p.Nearest(false, _Zebra)
	storage.Lock()
	if z == nil {
		// тот случай когда нет зебр и как заводить хозяйство не понятно
		log.Print("::WARN::")
		time.Sleep(12 * time.Second)
		return
	}
	zid := z.Id
	storage.Unlock()
	cb, zebra := storage.GetAnimalById(zid)
	cb()
	stage := 0
	for {
		select {
		case <-timer.C:
			if stage == 0 {
				// поймать зебру
				if dist(z, &p._BaseEntity) > TrapRadius {
					p.GoToObject(z, 0.3)
				} else {
					stage = 1
					MovingChannel <- zebra.Id
				}
			} else if stage == 1 {
				// подойти к ней
				p.GoToObject(&_BaseEntity{
					Left: zebra.Left - EntityWidth,
					Top:  zebra.Top,
				}, 0.3)
				if p.Left == zebra.Left-EntityWidth && p.Top == zebra.Top {
					stage = 2
				}
			} else if stage == 2 {
				p.GoToObject(&_BaseEntity{
					Top:  p.Farm.Top,
					Left: p.Farm.Left - EntityWidth,
				}, 0.3)
				zebra.GoToObject(&p._BaseEntity,
					float64(zebra.Hunger)/MaxPointLiveHunger)
				if p.Top == p.Farm.Top && p.Left == p.Farm.Left {
					// todo: сохранить животное или/и задать ей границы движимости
					return
				}
			}
		}
	}
}

func (p *Human) BuildFarm(l1 int, t1 int, l2 int, t2 int) (farm *Farm) {
	farm = &Farm{
		_BaseEntity: *GenerateBaseEntity(_Farm),
		fencing:     []*Fence{},
	}
	progress := false
	target := 1
	helper := func(l int, t int, newProgress bool, newTarget int) {
		p.GoToObject(&_BaseEntity{
			Top:  t1,
			Left: l1,
		}, 0.1)
		if p.Top == t && p.Left == l {
			progress, target = newProgress, newTarget
		}
	}
	building := func(l1 int, t1 int, l2 int, t2 int) {
		f := &Fence{begin: Point{
			Left: l1,
			Top:  t1,
		}, end: Point{
			Left: l2,
			Top:  t2,
		}, Id: getNextId()}
		farm.fencing = append(farm.fencing, f)
		storage.AddFence(f)
		WriteJSON(struct {
			OnCmd  Command
			Point1 Point
			Point2 Point
		}{
			OnCmd:  MakeFence,
			Point1: f.begin,
			Point2: f.end,
		})
	}
	t := time.NewTicker(LifeCyclePeriod)
	for {
		select {
		case <-t.C:
			if !progress {
				helper(l1, t1, true, 2)
			} else {
				switch target {
				case 2:
					helper(l1, t2, true, 3)
					if target == 3 {
						building(l1, t1, l1, t2)
					}
				case 3:
					helper(l2, t2, true, 4)
					if target == 4 {
						building(l1, t2, l2, t2)
					}
				case 4:
					helper(l2, t1, true, 5)
					if target == 5 {
						building(l2, t2, l2, t1)
					}
				case 5:
					helper(l1, t1, true, 6)
					if target == 6 {
						building(l2, t1, l1, t1)
					}
				case 6:
					return
				}
			}
		}
	}
}

func (h *House) Locate(gender Gender) *_BaseEntity {
	if gender == Male {
		return &_BaseEntity{
			Top:  h.Top + 30,
			Left: h.Left + 50,
		}
	} else {
		return &_BaseEntity{
			Top:  h.Top + 30,
			Left: h.Left + 20,
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

var CityMap = map[int]bool{}

func CityBuilder(es ...*_BaseEntity) (int, int) {
	const INF = int(1e9 + 7)
	var mnval, mni, mnj = INF, 0, 0

	for i := 0; i < AllWidth; i += HouseWidth {
		for j := 0; j < AllHeight; j += HouseHeight {
			if _, ok := CityMap[i*AllWidth+j]; ok {
				continue
			}
			var now = 0
			for _, k := range es {
				now += absForInt(k.Left-i) + absForInt(k.Top-j)
			}
			for ii := range []int{2, -1, 0, 0} {
				for jj := range []int{0, 0, 2, -1} {
					br := (i+ii*HouseWidth)*AllWidth + (j + jj*HouseHeight)
					if _, ok := CityMap[br]; ok {
						now -= FriendshipRatio
					}
				}
			}

			if mnval > now {
				mnval, mni, mnj = now, i, j
			}
		}
	}
	CityMap[mni*AllWidth+mnj] = true
	return mni, mnj
}

func (h *House) CreateChild() {
	c := &Human{
		_BaseAnimal: _BaseAnimal{
			_BaseEntity: *GenerateBaseEntity(
				_Human,
				h.Left+50,
				h.Top+50,
			),
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
	WriteJSON(c.asCmdCreate())
	time.AfterFunc(5*time.Second, func() {
		c.Age = 6
		c.SendChangeAge()
		c.SendMoveMe(0, 0, 0.8)
	})
	time.AfterFunc(10*time.Second, func() {
		c.Age = 19
		storage.AddHuman(c)
		c.Left -= 15
		c.SendMoveMe(-15, 0, 0.8)
		go c.LifeCycle()
	})
}

func dist(a *_BaseEntity, b *_BaseEntity) int {
	return absForInt(a.Left-b.Left) + absForInt(a.Top-b.Top)
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
		if skind, ok := storage.GetTypeById(*id); !ok {
			log.Fatalf("В программе баг")
		} else {
			kind = &skind
		}
	}
	return m[*kind] * (1 - 2*rand.Intn(2))
}

func (e *_BaseEntity) Nearest(noBlock bool, params ...TypeEntity) *_BaseEntity {
	return e.NearestLst(params, noBlock)
}

func (e *_BaseEntity) NearestLst(params []TypeEntity, noBlock bool) *_BaseEntity {
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
			if o == nil || e.Id == o.Id {
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
	for _, d := range storage.AllBaseEntities(noBlock) {
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
	for count := CountPeople; count > 0; count-- {
		curGender := Male
		if count <= 2 {
			curGender = Female
		}
		curAge := 19
		addPeople(curGender, curAge)
	}
}

var MovingChannel = make(chan int, 200)

func (c *Client) MovingManager() {
	ticker := time.NewTicker(MovingPeriod)
	all := map[int]*_BaseAnimal{}

	for !IsClosed(c.die) {
		select {
		case <-ticker.C:
			for id, o := range all {
				storage.Lock()
				if o == nil {
					delete(all, id)
					storage.Unlock()
					continue
				}

				t := o.NearestLst(o.Kind.GetTarget(), true)
				if t == nil {
					t = o.Nearest(true, o.Kind)
					if t == nil {
						// лол че делать (нет еды и это последняя особь своего вида)
						log.Printf("%s не знает чем заняться", o.String())
						continue
					}
				}
				o.GoToObject(t, float64(o.Hunger)/MaxPointLiveHunger)
				storage.Unlock()

			}
		case id := <-MovingChannel:
			if _, ok := all[id]; !ok {
				var cb func()
				cb, all[id] = storage.GetAnimalById(id)
				cb()
			} else {
				delete(all, id)
			}
		}
	}
}

func (p *_BaseAnimal) SendMoveMe(dirX int, dirY int, hunger float64) {
	/* Отправляет клиенту результат передвижения модели */
	WriteJSON(struct {
		OnCmd   Command
		ChangeX int
		ChangeY int
		IdObj   int
		Hunger  float64
	}{
		OnCmd:   MoveMe,
		ChangeX: dirX,
		ChangeY: dirY,
		IdObj:   p.Id,
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
			for _, obj1 := range storage.AllBaseEntities() {
				for _, obj2 := range storage.AllBaseEntities() {
					if obj1.Id == obj2.Id {
						continue
					}
					if !areMet(obj1.Left, obj1.Top, obj2.Left, obj2.Top) {
						continue
					}
					log.Printf("%s and %s met", obj1.String(), obj2.String())

					if obj1.Kind.isAnimal() && obj2.Kind.inLst(obj1.Kind.GetTarget()) {
						obj2.remove(Eaten)
						if cb, an := storage.GetAnimalById(obj1.Id); an != nil {
							func() {
								defer cb()
								an.Hunger -= RidPointHungerIfKill
								if an.Hunger < 0 {
									an.Hunger = 0
								}
							}()
						}
					}

					if obj1.Kind == _Human && obj2.Kind.inLst(obj1.Kind.GetTarget()) {
						log.Printf("obj2=%s (obj1=%s)", obj2.String(), obj1.String())
						obj2.remove(Eaten)
						h := storage.GetHumanById(obj1.Id)
						h.Take(TelegramMessage{Head: KillFood})
					}

					if obj1.Kind == obj2.Kind && obj1.Kind.isAnimal() {
						MakeChild(obj1, obj2)
					}
				}
			}
		}
	}
}

var NoInc = map[int]bool{}

func MakeChild(o1 *_BaseEntity, o2 *_BaseEntity) {
	if o1.Kind != o2.Kind {
		log.Panicf("Эта функция не может скрещивать %s и %s", o1.Kind, o2.Kind)
	}
	if !helper.AdderAnimalInitiate {
		log.Printf("Эта функция ожидает завершения генерации животных")
		return
	}
	_, ok1 := NoInc[o1.Id*EntitiesLimit+o2.Id]
	_, ok2 := NoInc[o2.Id*EntitiesLimit+o1.Id]
	if ok1 || ok2 {
		return
	}
	if l := len(storage.AllAnimal()); l > 18 {
		return
	}
	log.Print("make child!")

	if fn, ok := map[TypeEntity]func(entity *_BaseEntity) int{
		_Elephant: helper.AddElephant,
		_Bear:     helper.AddBear,
		_Fox:      helper.AddFox,
		_Rabbit:   helper.AddRabbit,
		_Wolf:     helper.AddWolf,
		_Zebra:    helper.AddZebra,
	}[o1.Kind]; ok {
		var newid = fn(GenerateBaseEntity(
			o1.Kind,
			(o1.Left+o2.Left)/2,
			(o1.Top+o2.Top)/2,
		))
		NoInc[o1.Id*EntitiesLimit+newid] = true
		NoInc[o2.Id*EntitiesLimit+newid] = true
		NoInc[o1.Id*EntitiesLimit+o2.Id] = true
	} else {
		log.Panicf("Эта функция не может скрещивать %s и %s", o1.Kind, o2.Kind)
	}
}

func (p *Human) SendChangeAge() {
	WriteJSON(struct {
		OnCmd  Command
		Id     int
		Age    int
		Gender Gender
	}{
		OnCmd:  ChangeAge,
		Id:     p.Id,
		Age:    p.Age,
		Gender: p.Gender,
	})
}

func (c *Client) Populate() {
	ticker := time.NewTicker(RandomPopulatingPeriod)

	for !IsClosed(c.die) {
		select {
		case <-ticker.C:
			if len(storage.AllPlants()) < 5 {
				addPlant()
			}
			if helper.AdderAnimalInitiate {
				if len(storage.AllAnimal()) < 18 {
					switch rand.Intn(9) {
					case 0:
						helper.AddZebra(GenerateBaseEntity(_Zebra))
					case 1:
						helper.AddWolf(GenerateBaseEntity(_Wolf))
					case 3:
						helper.AddFox(GenerateBaseEntity(_Fox))
					case 4:
						helper.AddBear(GenerateBaseEntity(_Bear))
					case 5:
						helper.AddElephant(GenerateBaseEntity(_Elephant))
					default: // вероятность рождения зайца больше - они нужнее
						helper.AddRabbit(GenerateBaseEntity(_Rabbit))
					}
				}
			}
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
	EntitiesLimit  = 1000
	CountElephants = 2
	CountFoxes     = 2
	CountBears     = 2
	CountWolfs     = 3
	CountZebras    = 2
	CountRabbits   = 15
	CountPlants    = 15
	CountPeople    = 4

	EntityWidth  = 30
	EntityHeight = 30
	HouseWidth   = 80
	HouseHeight  = 80

	RidPointHungerIfKill = 60
	TrapRadius           = 120

	AllWidth  = 1800
	AllHeight = 900

	MovingPeriod           = 1000 * time.Millisecond
	MeetingCheckerPeriod   = 500 * time.Millisecond
	StarveProcessPeriod    = 1500 * time.Millisecond
	LifeCyclePeriod        = 1000 * time.Millisecond
	RandomPopulatingPeriod = 5 * time.Second

	Child      SocialStatus = "Ребенок еще (рано пока)"
	InSearch   SocialStatus = "В активном поиске"
	InTheWay   SocialStatus = "По пути к своей половинке"
	InMarriage SocialStatus = "Женат / замужем"

	DrawPeople Command = "DrawPeople"
	DrawPlant  Command = "DrawPlant"
	InfoAbout  Command = "InfoAbout"
	DrawAnimal Command = "DrawAnimal"
	MoveMe     Command = "MoveMe"
	MustDie    Command = "MustDie"
	Bue        Command = "Bue"
	DrawHouse  Command = "DrawHouse"
	MakeFence  Command = "MakeFence"
	ChangeAge  Command = "ChangeAge"

	KillFood       HeadTelegramMessage = "Покушал"
	LetSGetMarried HeadTelegramMessage = "letSGetMarried"
	GoToState9     HeadTelegramMessage = "GoToState9"
	HouseHasBuilt  HeadTelegramMessage = "HouseHasBuilt"
	ImGoingAtHome  HeadTelegramMessage = "Покушал, иду домой"
	IAmAtHome      HeadTelegramMessage = "Я дома"

	Starvation       Reason = "Умер от голода"
	Eaten            Reason = "Его съели"
	LimitConnections Reason = "Maximum concurrent connections exceeded"

	MaxPointLiveHunger = 100
	FriendshipRatio    = 1000
)
