package main

import (
	"fmt"
	"sync"
)

type Storage struct {
	lock     sync.RWMutex
	ToType   map[int]TypeEntity

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
}

func NewStorage() (s *Storage) {
	return &Storage{
		lock:       sync.RWMutex{},
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
	}
}

func (s *Storage) addRabbit(o *Rabbit) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Rabbit
	s._rabbits[o.Id] = o
}
func (s *Storage) addWolf(o *Wolf) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Wolf
	s._wolfs[o.Id] = o
}
func (s *Storage) addHuman(o *Human) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Human
	s._people[o.Id] = o
}
func (s *Storage) addHouse(o *House) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _House
	s._houses[o.Id] = o
}
func (s *Storage) addBear(o *Bear) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Bear
	s._bears[o.Id] = o
}
func (s *Storage) addZebra(o *Zebra) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Zebra
	s._zebras[o.Id] = o
}
func (s *Storage) addFox(o *Fox) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Fox
	s._foxes[o.Id] = o
}
func (s *Storage) addCabbage(o *Cabbage) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Cabbage
	s._cabbage[o.Id] = o
}
func (s *Storage) addBush(o *Bush) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Bush
	s._bushes[o.Id] = o
}
func (s *Storage) addCarrot(o *Carrot) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Carrot
	s._carrots[o.Id] = o
}
func (s *Storage) addElephant(o *Elephant) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ToType[o.Id] = _Elephant
	s._elephants[o.Id] = o
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
)

func (t TypeEntity) String() string {
	return [...]string{"Кролик", "Волк", "Человек", "Дом",
		"Медведь", "Зебра", "Лиса", "Капуста", "Кустарник",
		"Морковь", "Слон"}[t]
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

func (e *_BaseEntity) String() string {
	return fmt.Sprintf("%s#%d", e.Kind.String(), e.Id)
}

type _BasePlant struct {
	_BaseEntity
}

func (s *Storage) GetAnimalById(id int) (cb func(), an *_BaseAnimal) {
	s.lock.Lock()
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
	return
}

func (s *Storage) GetPlantById(id int) *_BasePlant {
	s.lock.Lock()
	defer s.lock.Unlock()
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

func (s *Storage) GetHumanById(id int) *Human {
	s.lock.Lock()
	defer s.lock.Unlock()
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

func (s *Storage) GetHouseById(id int) *House {
	s.lock.Lock()
	defer s.lock.Unlock()
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

func (s *Storage) ExistId(id int) (ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, ok = s.ToType[id]
	return
}

func (s *Storage) AllPlants() (m map[int]*_BasePlant) {
	s.lock.Lock()
	defer s.lock.Unlock()
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

func (s *Storage) AllAnimal() (m map[int]*_BaseAnimal) {
	s.lock.Lock()
	defer s.lock.Unlock()
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

func (s *Storage) AllPeople() (m map[int]*Human) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s._people
}

func (s *Storage) AllHouses() (m map[int]*House) {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s._houses
}

func (s *Storage) AllBaseEntities() (m map[int]*_BaseEntity) {
	s.lock.Lock()
	defer s.lock.Unlock()
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
	return
}

func (s *Storage) RemoveById(id int, blocking ... bool) {
	if len(blocking) == 0 || blocking[0] {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	// todo: fix this
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

type Gender string
