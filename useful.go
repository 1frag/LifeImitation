package main

func IsClosed(ch <-chan bool) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

type Helper struct {
	AddRabbit           func(entity *_BaseEntity) int
	AddZebra            func(entity *_BaseEntity) int
	AddWolf             func(entity *_BaseEntity) int
	AddBear             func(entity *_BaseEntity) int
	AddFox              func(entity *_BaseEntity) int
	AddElephant         func(entity *_BaseEntity) int
	AddCarrot           func(entity *_BaseEntity) int
	AddCabbage          func(entity *_BaseEntity) int
	AddBush             func(entity *_BaseEntity) int
	AdderAnimalInitiate bool
}
