package main

import (
	"fmt"
	"sync"
)

type Conf string

const (
	NoWait Conf = "NoWait"
	Wait Conf = "Wait"
)

type Storage struct {
	lock   *CustomMutex
	ToType map[int]TypeEntity

	_rabbits   map[int]*Rabbit
	_wolfs     map[int]*Wolf
	_bears     map[int]*Bear
	_zebras    map[int]*Zebra
	_foxes     map[int]*Fox
	_people    map[int]*Human
	_houses    map[int]*House
	_cabbage   map[int]*Cabbage
	_bushes    map[int]*Bush
	_carrots   map[int]*Carrot
	_elephants map[int]*Elephant
	_farms     map[int]*Farm
	_fences    map[int]*Fence
}

func NewStorage() (s *Storage) {
	return &Storage{
		lock:       NewCustomMutex(),
		ToType:     map[int]TypeEntity{},
		_rabbits:   map[int]*Rabbit{},
		_wolfs:     map[int]*Wolf{},
		_bears:     map[int]*Bear{},
		_zebras:    map[int]*Zebra{},
		_foxes:     map[int]*Fox{},
		_people:    map[int]*Human{},
		_houses:    map[int]*House{},
		_cabbage:   map[int]*Cabbage{},
		_bushes:    map[int]*Bush{},
		_carrots:   map[int]*Carrot{},
		_elephants: map[int]*Elephant{},
		_farms:     map[int]*Farm{},
		_fences:    map[int]*Fence{},
	}
}

func NewCustomMutex() *CustomMutex {
	return &CustomMutex{
		mutex: sync.RWMutex{},
	}
}

type CustomMutex struct {
	mutex sync.RWMutex
}

func (m *CustomMutex) Lock(v ...Conf) {
	if len(v) == 1 && v[0] == NoWait {
		return
	}
	//log.Printf("\n\n\nStack: %s", string(debug.Stack()))
	m.mutex.Lock()
}
func (m *CustomMutex) Unlock(v ...Conf) {
	if len(v) == 1 && v[0] == NoWait {
		return
	}
	//log.Print("Release!\n\n\n")
	m.mutex.Unlock()
}

func (s *Storage) AddRabbit(o *Rabbit, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Rabbit
	s._rabbits[o.Id] = o
}
func (s *Storage) AddWolf(o *Wolf, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Wolf
	s._wolfs[o.Id] = o
}
func (s *Storage) AddHuman(o *Human, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Human
	s._people[o.Id] = o
}
func (s *Storage) AddHouse(o *House, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _House
	s._houses[o.Id] = o
}
func (s *Storage) AddBear(o *Bear, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Bear
	s._bears[o.Id] = o
}
func (s *Storage) AddZebra(o *Zebra, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Zebra
	s._zebras[o.Id] = o
}
func (s *Storage) AddFox(o *Fox, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Fox
	s._foxes[o.Id] = o
}
func (s *Storage) AddCabbage(o *Cabbage, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Cabbage
	s._cabbage[o.Id] = o
}
func (s *Storage) AddBush(o *Bush, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Bush
	s._bushes[o.Id] = o
}
func (s *Storage) AddCarrot(o *Carrot, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Carrot
	s._carrots[o.Id] = o
}
func (s *Storage) AddElephant(o *Elephant, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Elephant
	s._elephants[o.Id] = o
}
func (s *Storage) AddFarm(o *Farm, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Farm
	s._farms[o.Id] = o
}
func (s *Storage) AddFence(o *Fence, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	s.ToType[o.Id] = _Fence
	s._fences[o.Id] = o
}

type TypeEntity int

const (
	_Rabbit = iota
	_Wolf
	_Human
	_House
	_Bear
	_Zebra
	_Fox
	_Cabbage
	_Bush
	_Carrot
	_Elephant
	_Fence
	_Farm
)

func (t TypeEntity) String() string {
	return [...]string{"Кролик", "Волк", "Человек", "Дом",
		"Медведь", "Зебра", "Лиса", "Капуста", "Кустарник",
		"Морковь", "Слон", "Забор", "Ферма"}[t]
}

type Cabbage struct{ _BasePlant }
type Bush struct{ _BasePlant }
type Carrot struct{ _BasePlant }
type Rabbit struct{ _BaseAnimal }   // кушает: капусту и морковку
type Zebra struct{ _BaseAnimal }    // кушает: кусты
type Elephant struct{ _BaseAnimal } // кушает: все растения
type Fox struct{ _BaseAnimal }      // кушает: кроликов
type Wolf struct{ _BaseAnimal }     // кушает: кроликов и лисиц
type Bear struct{ _BaseAnimal }     // кушает: кроликов, лисиц, волков и зебр
type House struct {
	_BaseEntity
	Wife    *Human
	Husband *Human
}
type Human struct {
	_BaseAnimal
	Age          int
	State        int
	Target       *Human
	Gender       Gender
	SocialStatus SocialStatus
	Telegram     chan TelegramMessage
	House        *House
	Farm         *Farm
	Pets         []*Pet
}
type Point struct {
	Left int
	Top  int
}
type Fence struct {
	Id    int
	begin Point
	end   Point
}
type Farm struct {
	_BaseEntity
	fencing []*Fence
}

type _BaseAnimal struct {
	_BaseEntity
	Hunger int
	Target *_BaseEntity
}

type _BaseEntity struct {
	Id   int
	Top  int
	Left int
	Kind TypeEntity
	die  chan bool
}

type Pet struct {
	entity  *_BaseEntity
}

func (e *_BaseEntity) String() string {
	return fmt.Sprintf("%s#%d", e.Kind.String(), e.Id)
}

type _BasePlant struct {
	_BaseEntity
}

func (s *Storage) GetTypeById(id int, v ...Conf) (t TypeEntity, b bool) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	t, b = s.ToType[id]
	return
}

func (s *Storage) Lock(v ...Conf) {
	s.lock.Lock(v...)
}

func (s *Storage) Unlock(v ...Conf) {
	s.lock.Unlock(v...)
}

func (s *Storage) GetAnimalById(id int, v ...Conf) (cb func(v ...Conf), an *_BaseAnimal) {
	s.lock.Lock(v...)
	cb = s.lock.Unlock
	t, ok := s.ToType[id]
	if !ok {
		cb()
		return nil, nil
	}
	switch t {
	case _Rabbit:
		an = &s._rabbits[id]._BaseAnimal
	case _Bear:
		an = &s._bears[id]._BaseAnimal
	case _Elephant:
		an = &s._elephants[id]._BaseAnimal
	case _Fox:
		an = &s._foxes[id]._BaseAnimal
	case _Zebra:
		an = &s._zebras[id]._BaseAnimal
	case _Wolf:
		an = &s._wolfs[id]._BaseAnimal
	}
	if an == nil {
		cb()
	}
	return
}

func (s *Storage) GetPlantById(id int, v ...Conf) *_BasePlant {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	t, ok := s.ToType[id]
	if !ok {
		return nil
	}
	switch t {
	case _Bush:
		return &s._bushes[id]._BasePlant
	case _Cabbage:
		return &s._cabbage[id]._BasePlant
	case _Carrot:
		return &s._carrots[id]._BasePlant
	}
	return nil
}

func (s *Storage) GetHumanById(id int, v ...Conf) *Human {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	t, ok := s.ToType[id]
	if !ok {
		return nil
	}
	switch t {
	case _Human:
		return s._people[id]
	}
	return nil
}

func (s *Storage) GetHouseById(id int, v ...Conf) *House {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	t, ok := s.ToType[id]
	if !ok {
		return nil
	}
	switch t {
	case _House:
		return s._houses[id]
	}
	return nil
}

func (s *Storage) ExistId(id int, v ...Conf) (ok bool) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	_, ok = s.ToType[id]
	return
}

func (s *Storage) AllPlants(v ...Conf) (m map[int]*_BasePlant) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	m = make(map[int]*_BasePlant)
	for i, d := range s._carrots {
		m[i] = &d._BasePlant
	}
	for i, d := range s._cabbage {
		m[i] = &d._BasePlant
	}
	for i, d := range s._bushes {
		m[i] = &d._BasePlant
	}
	return
}

func (s *Storage) AllAnimal(v ...Conf) (m map[int]*_BaseAnimal) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	m = make(map[int]*_BaseAnimal)
	for i, d := range s._rabbits {
		m[i] = &d._BaseAnimal
	}
	for i, d := range s._bears {
		m[i] = &d._BaseAnimal
	}
	for i, d := range s._elephants {
		m[i] = &d._BaseAnimal
	}
	for i, d := range s._foxes {
		m[i] = &d._BaseAnimal
	}
	for i, d := range s._zebras {
		m[i] = &d._BaseAnimal
	}
	for i, d := range s._wolfs {
		m[i] = &d._BaseAnimal
	}
	return
}

func (s *Storage) AllPeople(v ...Conf) (m map[int]*Human) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	return s._people
}

func (s *Storage) AllHouses(v ...Conf) (m map[int]*House) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	return s._houses
}

func (s *Storage) AllFarms(v ...Conf) (m map[int]*Farm) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)
	return s._farms
}

func (s *Storage) AllBaseEntities(v ...Conf) (m map[int]*_BaseEntity) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)

	m = map[int]*_BaseEntity{}
	for i, d := range s._bushes {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._cabbage {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._carrots {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._wolfs {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._zebras {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._foxes {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._people {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._houses {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._elephants {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._bears {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._rabbits {
		m[i] = &d._BaseEntity
	}
	for i, d := range s._farms {
		m[i] = &d._BaseEntity
	}
	return
}

func (s *Storage) RemoveById(id int, v ...Conf) {
	s.lock.Lock(v...)
	defer s.lock.Unlock(v...)

	delete(s.ToType, id)
	delete(s._bushes, id)
	delete(s._cabbage, id)
	delete(s._carrots, id)
	delete(s._wolfs, id)
	delete(s._zebras, id)
	delete(s._foxes, id)
	delete(s._people, id)
	delete(s._houses, id)
	delete(s._elephants, id)
	delete(s._bears, id)
	delete(s._rabbits, id)
}

func (t TypeEntity) GetTarget() []TypeEntity {
	return map[TypeEntity][]TypeEntity{
		_Zebra:    {_Bush},
		_Rabbit:   {_Carrot, _Cabbage},
		_Elephant: {_Bush},
		_Fox:      {_Rabbit},
		_Wolf:     {_Rabbit, _Fox},
		_Bear:     {_Rabbit, _Fox, _Wolf, _Zebra},
		_Human:    {_Rabbit, _Wolf, _Cabbage, _Carrot},
	}[t]
}

func (t TypeEntity) isAnimal() bool {
	return t.in(_Zebra, _Wolf, _Bear, _Fox, _Rabbit, _Elephant)
}

func (t TypeEntity) in(l ...TypeEntity) bool {
	for _, e := range l {
		if e == t {
			return true
		}
	}
	return false
}

func (t TypeEntity) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.String())), nil
}

func (p *Point) String() string {
	return fmt.Sprintf("(%d, %d)", p.Left, p.Top)
}

func (f *Fence) String() string {
	return fmt.Sprintf("%s -> %s", f.begin.String(), f.end.String())
}

type Gender string
func (g Gender) String() string {
	if g == Male {
		return "Male"
	} else {
		return "Female"
	}
}
