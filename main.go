// file: main.go
package main

import (
	"fmt"
	"sort"
	"strings"
)

type Person struct {
	BackPack  bool
	Inventory map[string]int
}

type Game struct {
	Current  *Area
	Areas    map[string]*Area
	Player   Person
	DoorOpen bool
}

type Area struct {
	Name        string
	Description string
	Neighbors   map[string]*Area
	Objects     map[string]string
	UseActions  map[string]map[string]func(*Game) string
}

var game Game

func initGame() {
	kitchen := &Area{
		Name:        "кухня",
		Description: "кухня, ничего интересного",
		Objects:     map[string]string{"чай": "стол"},
	}
	corridor := &Area{
		Name:        "коридор",
		Description: "ничего интересного",
		Objects:     map[string]string{},
	}
	room := &Area{
		Name:        "комната",
		Description: "ты в своей комнате",
		Objects: map[string]string{
			"ключи":     "стол",
			"конспекты": "стол",
			"рюкзак":    "стул",
		},
	}
	street := &Area{
		Name:        "улица",
		Description: "на улице весна",
		Objects:     map[string]string{},
	}

	kitchen.Neighbors = map[string]*Area{"коридор": corridor}
	corridor.Neighbors = map[string]*Area{
		"кухня":   kitchen,
		"комната": room,
		"улица":   street,
	}
	room.Neighbors = map[string]*Area{"коридор": corridor}
	street.Neighbors = map[string]*Area{"домой": corridor}

	for _, a := range []*Area{kitchen, corridor, room, street} {
		a.UseActions = make(map[string]map[string]func(*Game) string)
	}
	corridor.UseActions["ключи"] = map[string]func(*Game) string{
		"дверь": func(g *Game) string {
			if !g.DoorOpen {
				g.DoorOpen = true
				return "дверь открыта"
			}
			return "дверь уже открыта"
		},
	}

	game = Game{
		Current:  kitchen,
		Areas:    map[string]*Area{"кухня": kitchen, "коридор": corridor, "комната": room, "улица": street},
		Player:   Person{BackPack: false, Inventory: make(map[string]int)},
		DoorOpen: false,
	}
}

func handleCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "неизвестная команда"
	}
	switch parts[0] {
	case "осмотреться":
		return cmdLook()
	case "идти":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return cmdGo(parts[1])
	case "надеть":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return cmdWear(parts[1])
	case "взять":
		if len(parts) < 2 {
			return "неизвестная команда"
		}
		return cmdTake(parts[1])
	case "применить":
		if len(parts) < 3 {
			return "неизвестная команда"
		}
		return cmdUse(parts[1], parts[2])
	default:
		return "неизвестная команда"
	}
}

func cmdLook() string {
	area := game.Current
	switch area.Name {
	case "кухня":
		places := describePlaces(area.Objects)
		suffix := "надо идти в универ."
		if !game.Player.BackPack {
			suffix = "надо собрать рюкзак и идти в универ."
		}
		return fmt.Sprintf(
			"ты находишься на кухне, %s, %s можно пройти - %s",
			places, suffix, neighborsList(area),
		)
	case "комната":
		if len(area.Objects) == 0 {
			return fmt.Sprintf(
				"пустая комната. можно пройти - %s",
				neighborsList(area),
			)
		}
		places := describePlaces(area.Objects)
		return fmt.Sprintf(
			"%s. можно пройти - %s",
			places, neighborsList(area),
		)
	default:
		return fmt.Sprintf(
			"ничего интересного. можно пройти - %s",
			neighborsList(area),
		)
	}
}

func cmdGo(dest string) string {
	next, ok := game.Current.Neighbors[dest]
	if !ok {
		return "нет пути в " + dest
	}
	if dest == "улица" && !game.DoorOpen {
		return "дверь закрыта"
	}
	game.Current = next
	return fmt.Sprintf(
		"%s. можно пройти - %s",
		next.Description,
		neighborsList(next),
	)
}

func cmdWear(item string) string {
	area := game.Current
	if item == "рюкзак" {
		if _, ok := area.Objects["рюкзак"]; ok {
			game.Player.BackPack = true
			delete(area.Objects, "рюкзак")
			return "вы надели: рюкзак"
		}
	}
	return "нет такого"
}

func cmdTake(item string) string {
	area := game.Current
	if !game.Player.BackPack {
		return "некуда класть"
	}
	if _, ok := area.Objects[item]; !ok {
		return "нет такого"
	}
	game.Player.Inventory[item]++
	delete(area.Objects, item)
	return "предмет добавлен в инвентарь: " + item
}

func cmdUse(item, target string) string {
	if game.Player.Inventory[item] == 0 {
		return "нет предмета в инвентаре - " + item
	}
	area := game.Current
	if acts, ok := area.UseActions[item]; ok {
		if action, ok2 := acts[target]; ok2 {
			return action(&game)
		}
	}
	return "не к чему применить"
}

func describePlaces(objects map[string]string) string {
	loc := map[string]string{
		"стол": "столе",
		"стул": "стуле",
	}
	byPlace := make(map[string][]string, len(objects))
	for obj, place := range objects {
		byPlace[place] = append(byPlace[place], obj)
	}
	places := make([]string, 0, len(byPlace))
	for p := range byPlace {
		places = append(places, p)
	}
	sort.Strings(places)

	var parts []string
	for _, p := range places {
		items := byPlace[p]
		sort.Strings(items)
		name := p
		if v, ok := loc[p]; ok {
			name = v
		}
		parts = append(parts, fmt.Sprintf("на %s: %s", name, strings.Join(items, ", ")))
	}
	return strings.Join(parts, ", ")
}

func neighborsList(area *Area) string {
	keys := make([]string, 0, len(area.Neighbors))
	for k := range area.Neighbors {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func main() {
}
