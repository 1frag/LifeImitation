package main

import "C"
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

func GenerateBaseEntity() *BaseEntity {
	return &BaseEntity{
		Id:   getNextId(),
		Top:  rand.Intn(AllHeight),
		Left: rand.Intn(AllWidth),
		die:  make(chan bool, 2),
	}
}

func addPlant() {
	pl := &Plant{
		BaseEntity: *GenerateBaseEntity(),
		Type:       rand.Intn(6),
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
var StoragePeople MapOfPeople = make(map[int]*People)
var StorageHouses MapOfHouses = make(map[int]*House)

type MapOfPlants map[int]*Plant
type MapOfHAnimal map[int]*HerbivoreAnimal
type MapOfPAnimal map[int]*PredatoryAnimal
type MapOfBEntity map[int]*BaseEntity
type MapOfPeople map[int]*People
type MapOfHouses map[int]*House

// todo: how to fix this shit?
func (s *MapOfPlants) getBaseEntity() (r MapOfBEntity) {
	r = make(map[int]*BaseEntity)
	for i, e := range *s {
		r[i] = &e.BaseEntity
	}
	return
}

func (s *MapOfPeople) getBaseEntity() (r MapOfBEntity) {
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

func (s *MapOfHouses) getBaseEntity() (r MapOfBEntity) {
	r = make(map[int]*BaseEntity)
	for i, e := range *s {
		r[i] = &e.BaseEntity
	}
	return
}

func _(id int) *BaseEntity {
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

func GetMovEntity(id int) *BaseAnimal {
	if o, ok := StoragePredatoryAnimal[id]; ok {
		return &o.BaseAnimal
	}
	if o, ok := StorageHerbivoreAnimal[id]; ok {
		return &o.BaseAnimal
	}
	return nil
}

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

func (e *BaseEntity) remove(reason Reason) {
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

func (p *Plant) remove(reason Reason) {
	if p != nil {
		p.BaseEntity.remove(reason)
		delete(StoragePlants, p.Id)
	}
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

func (p *People) remove(reason Reason) {
	if p != nil {
		p.BaseEntity.remove(reason)
		delete(StoragePeople, p.Id)
	}
}

type JSONForDrawEntity struct {
	OnCmd Command
	Top   int
	Left  int
	Type  int
	Id    int
}

func (p *Plant) AsCmdToJs() []byte {
	data, err := json.Marshal(JSONForDrawEntity{
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
				BaseEntity: *GenerateBaseEntity(),
				Hunger:     0,
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
	var data = make(map[int]map[string]string)
	for i, o := range StoragePlants {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		about["type"] = string(o.Type)
		data[i] = about
	}
	for i, o := range StorageHerbivoreAnimal {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		about["hunger"] = string(o.Hunger)
		data[i] = about
	}
	for i, o := range StoragePredatoryAnimal {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		about["hunger"] = string(o.Hunger)
		data[i] = about
	}
	for i, o := range StoragePeople {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		about["hunger"] = string(o.Hunger)
		if o.Target == nil {
			about["target"] = "__nil__"
		} else {
			about["target"] = string(o.Target.Id)
		}
		data[i] = about
	}
	for i, o := range StoragePeople {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		about["hunger"] = string(o.Hunger)
		if o.Target == nil {
			about["target"] = "__nil__"
		} else {
			about["target"] = string(o.Target.Id)
		}
		about["state"] = string(o.State)
		data[i] = about
	}
	for i, o := range StorageHouses {
		var about = make(map[string]string)
		about["left"] = string(o.Left)
		about["top"] = string(o.Top)
		if o.Husband == nil {
			about["Husband"] = "__nil__"
		} else {
			about["Husband"] = string(o.Husband.Id)
		}
		if o.Wife == nil {
			about["Wife"] = "__nil__"
		} else {
			about["Wife"] = string(o.Wife.Id)
		}
		data[i] = about
	}
	d, _ := json.Marshal(data)
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
	// todo: people and house
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
			if p.Hunger >= MaxPointLiveHunger {
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

type Gender string

const Male Gender = "Male"
const Female Gender = "Female"

type People struct {
	BaseAnimal
	Age          int
	State        int
	Target       *People
	Gender       Gender
	SocialStatus SocialStatus
	Telegram     chan TelegramMessage
	House        *House
}

type House struct {
	BaseEntity
	Wife    *People
	Husband *People
}

func addPeople(g Gender, a int) {
	p := &People{
		BaseAnimal: BaseAnimal{
			BaseEntity: *GenerateBaseEntity(),
			Hunger:     0,
			Target:     nil,
		},
		Age:      a,
		State:    0,
		Gender:   g,
		Telegram: make(chan TelegramMessage, 20),
	}
	writeJSON(p.asCmdCreate())
	StoragePeople[p.Id] = p
	go p.LifeCycle()
}

func (p *People) LifeCycle() {
	goToObject := func(target *BaseEntity, hunger float64) {
		l, t := target.Left, target.Top
		var dl = approach(p.Left, l, 5)
		var dt = approach(p.Top, t, 5)
		p.Left += dl
		p.Top += dt
		SendMoveMe(dl, dt, p.Id, hunger)
	}

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
					if m.Head == KillPlant {
						p.State += 1
					} else if m.Head == KillHAnimal {
						p.State += 2
					} else if m.Head == KillPAnimal {
						p.State += 3
					} else {
						log.Fatal("Unexpected Head Message")
					}
					if p.State > 6 {
						p.State = 7
					}
				default:
					// найди ближайшего и иди к нему
					goToObject(p.nearest(7), (7.0-float64(p.State))/8.0)
				}
			} else if p.State == 7 {
				log.Printf("%d has p.State==7 G=%s S=%s", p.Id, p.Gender, p.SocialStatus)
				// ищет себе вторую половинку
				onMet := func(o *People) {
					goToObject(&o.BaseEntity, 0)
					p.SocialStatus = InTheWay
					p.Target = o
					p.State = 8
				}

				p.SocialStatus = InSearch
				for _, o := range StoragePeople {
					if o.Gender != p.Gender && o.SocialStatus == InSearch {
						o.Telegram <- p.MessageWithSign(LetSGetMarried, nil)
						onMet(o)
						break
					}
				}
				if p.Target == nil {
					log.Printf("(%d) Стою без дела никого не нашел", p.Id)
					goToObject(p.nearest(7), 0)
				}
				select {
				case m := <-p.Telegram:
					if m.Head == LetSGetMarried {
						if pTarget, ok := StoragePeople[m.From]; ok {
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
				if _, ok := StoragePeople[p.Target.Id]; !ok {
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
					goToObject(&p.Target.BaseEntity, 0)
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
						if dist(&p.BaseEntity, &home.BaseEntity) < 50 {
							// дом построен
							writeJSON(JSONForDrawEntity{
								OnCmd: DrawHouse,
								Top:   home.Top,
								Left:  home.Left,
								Id:    home.Id,
							})
							StorageHouses[home.Id] = home
							p.Target.Telegram <- p.MessageWithSign(HouseHasBuilt, home.Id)
							p.State = 10
							p.House = home
						} else {
							// далеко, -> подойди поближе к дому чтобы его построить
							goToObject(&home.BaseEntity, 0)
						}
					} else {
						// А где строить то?
						// todo: 400, 50, 30
						h := p.nearest(16)
						var newHouse = House{
							BaseEntity: *GenerateBaseEntity(),
							Wife:       p.Target,
							Husband:    p,
						}
						nearby := &p.BaseEntity
						if h != nil && dist(h, &p.BaseEntity) < 400 {
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
								p.House, ok = StorageHouses[i]
							}
							if !ok {
								log.Fatal("Это никогда не должно произойти")
							}
							p.State = 10
						}
					default:
						goToObject(&BaseEntity{
							Top:  p.Target.Left,
							Left: p.Target.Top - EntityWidth,
						}, 0)
					}
				}
			} else if p.State == 10 {
				select {
				case m := <-p.Telegram:
					if m.Head == KillPlant || m.Head == KillHAnimal || m.Head == KillPAnimal {
						p.State = 11
						p.Target.Telegram <- p.MessageWithSign(ImGoingAtHome, nil)
					} else if m.Head == ImGoingAtHome || m.Head == IAmAtHome {
						partnerAtHome = true
					} else {
						log.Printf("Unexpected this head=(%s), target.id=%d," +
							" from.id=%d, pah=%t", m.Head, p.Target.Id,
							m.From, partnerAtHome)
					}
				default:
					goToObject(p.nearest(7), 0.3)
				}
			} else if p.State == 11 {
				if dist(&p.BaseEntity, p.House.Locate(p.Gender)) < 5 {
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
				break
			}
		}
	}
	log.Printf("The end for person with id=%d", p.Id)
}

func (h *House) Locate(gender Gender) *BaseEntity {
	if gender == Male {
		return &BaseEntity{
			Top:  h.Left + 50,
			Left: h.Top + 30,
		}
	} else {
		return &BaseEntity{
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
	c := &People{
		BaseAnimal: BaseAnimal{
			BaseEntity: BaseEntity{
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
		SendMoveMe(-15, 0, c.Id, 0.8)
	})
	time.AfterFunc(40*time.Second, func() {
		SendMoveMe(-15, 0, c.Id, 0.8)
	})
}

func dist2(a *BaseEntity, b *BaseEntity) float64 {
	return math.Pow(float64(a.Left-b.Left), 2) +
		math.Pow(float64(a.Top-b.Top), 2)
}

func dist(a *BaseEntity, b *BaseEntity) float64 {
	return math.Pow(dist2(a, b), 0.5)
}

func (p *People) MessageWithSign(header HeadTelegramMessage, body interface{}) TelegramMessage {
	return TelegramMessage{
		Head: header,
		From: p.Id,
		Body: body,
	}
}

func approach(from int, to int, rightSide int) int {
	if fv := func() int {
		if from < to {
			return randRange(0, rightSide)
		} else if from == to {
			return 0
		} else {
			return -randRange(0, rightSide)
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

func (p *People) Take(m TelegramMessage) {
	select {
	case p.Telegram <- m:
	default:
	}
}

func (e *BaseEntity) nearest(param int) *BaseEntity {
	/*
		1 - трава
		2 - зайцы
		4 - волки
		8 - люди
		16 - дома
	*/
	inner := func(s MapOfBEntity) *BaseEntity {
		bi, n := -1, -1
		var best *BaseEntity = nil
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
	var m MapOfBEntity = make(map[int]*BaseEntity)
	if param&1 == 1 {
		m[0] = inner(StoragePlants.getBaseEntity())
	}
	if param&2 == 2 {
		m[1] = inner(StoragePredatoryAnimal.getBaseEntity())
	}
	if param&4 == 4 {
		m[2] = inner(StorageHerbivoreAnimal.getBaseEntity())
	}
	if param&8 == 8 {
		m[3] = inner(StoragePeople.getBaseEntity())
	}
	if param&16 == 16 {
		m[4] = inner(StorageHouses.getBaseEntity())
	}
	return inner(m)
}

func (p *People) asCmdCreate() interface{} {
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

func GeneratePeoples() {
	for count := randRange(5, 10); count > 0; count-- {
		curGender := Male
		if rand.Intn(2) == 1 {
			curGender = Female
		}
		curAge := randRange(10, 40)
		addPeople(curGender, curAge)
	}
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
		Obj       *BaseAnimal
		Targeting string
	}

	var SafelyStep = func(o *BaseAnimal, dx int, dy int) (nl int, nt int) {
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
			o := GetMovEntity(id)
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

func (p *People) SendMoveMe(dirX int, dirY int, hunger float64, age int) {
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
			/* Травоядные животные и растения * /
			forr(StorageHerbivoreAnimal.getBaseEntity(),
				StoragePlants.getBaseEntity(), func(pid int, wid int) {
					StoragePlants[wid].remove(Eaten)
					StorageHerbivoreAnimal[pid].Hunger -= RidPointHungerIfKill
					if StorageHerbivoreAnimal[pid].Hunger < 0 {
						StorageHerbivoreAnimal[pid].Hunger = 0
					}
				})
			/* Хищные животные и травоядные * /
			forr(StoragePredatoryAnimal.getBaseEntity(),
				StorageHerbivoreAnimal.getBaseEntity(), func(pid int, wid int) {
					StorageHerbivoreAnimal[wid].remove(Eaten)
					StoragePredatoryAnimal[pid].Hunger -= RidPointHungerIfKill
					if StoragePredatoryAnimal[pid].Hunger < 0 {
						StoragePredatoryAnimal[pid].Hunger = 0
					}
				})*/

			/* Люди и трава */
			forr(StoragePeople.getBaseEntity(),
				StoragePlants.getBaseEntity(), func(pid int, wid int) {
					StoragePlants[wid].remove(Eaten)
					StoragePeople[pid].Take(TelegramMessage{Head: KillPlant})
				})

			/* Люди и зайцы */
			forr(StoragePeople.getBaseEntity(),
				StorageHerbivoreAnimal.getBaseEntity(), func(pid int, wid int) {
					StorageHerbivoreAnimal[wid].remove(Eaten)
					StoragePeople[pid].Take(TelegramMessage{Head: KillHAnimal})
				})

			/* Люди и волки */
			forr(StoragePeople.getBaseEntity(),
				StoragePredatoryAnimal.getBaseEntity(), func(pid int, wid int) {
					StoragePredatoryAnimal[wid].remove(Eaten)
					StoragePeople[pid].Take(TelegramMessage{Head: KillPAnimal})
				})
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
				BaseEntity: *GenerateBaseEntity(),
				Hunger:     0,
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

	PanelWidth  = 10
	PanelHeight = 10

	EntityWidth  = 30
	EntityHeight = 30

	AllWidth  = CountX*PanelWidth - EntityWidth
	AllHeight = CountY*PanelHeight - EntityHeight

	MovingPeriod        = 1000 * time.Millisecond
	KillCheckerPeriod   = 500 * time.Millisecond
	StarveProcessPeriod = 500 * time.Millisecond
	LifeCyclePeriod     = 1000 * time.Millisecond

	Child      SocialStatus = "Ребенок еще (рано пока)"
	InSearch   SocialStatus = "В активном поиске"
	InTheWay   SocialStatus = "По пути к своей половинке"
	InMarriage SocialStatus = "Женат / замужем"

	//DrawMapCmd          Command = "DrawMapCmd"
	DrawPeople          Command = "DrawPeople"
	DrawPlant           Command = "DrawPlant"
	InfoAbout           Command = "InfoAbout"
	DrawHerbivoreAnimal Command = "DrawHerbivoreAnimal"
	MoveMe              Command = "MoveMe"
	MustDie             Command = "MustDie"
	DrawPredatoryAnimal Command = "DrawPredatoryAnimal"
	Bue                 Command = "Bue"
	DrawHouse           Command = "DrawHouse"

	KillPlant      HeadTelegramMessage = "Kill->Plant"
	KillHAnimal    HeadTelegramMessage = "Kill->HAnimal"
	KillPAnimal    HeadTelegramMessage = "Kill->PAnimal"
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
	MinCountPlants     = 1
	MaxCountPlants     = 5
	MinCountHAnimal    = 10
	MaxCountHAnimal    = 12
	MinCountPAnimal    = 10
	MaxCountPAnimal    = 12
)
