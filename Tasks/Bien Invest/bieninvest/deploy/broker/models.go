package main

type User struct {
	Id        uint32
	Name      string
	Password  string
	Money     uint32
	Portfolio map[string]uint32
}
