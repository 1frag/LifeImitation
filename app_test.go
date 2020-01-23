package main

import "testing"

func TestApp(t *testing.T) {
	left, top := CityBuilder(&_BaseEntity{
		Top:  91,
		Left: 76,
		die:  nil,
	})
	t.Log(left, top)
	left, top = CityBuilder(&_BaseEntity{
		Top:  160,
		Left: 160,
		die:  nil,
	})
	t.Log(left, top)
}
