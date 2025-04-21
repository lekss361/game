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
	Current  string
	Areas    map[string]*Area
	Player   Person
	DoorOpen bool
}

type Area struct {
	Name      string
	Neighbors []string
	Objects   map[string]string
}

var game Game

func initGame() {
	game = Game{
		Current: "кухня",
		Player: Person{
			BackPack:  false,
			Inventory: make(map[string]int),
		},
		DoorOpen: false,
		Areas:    make(map[string]*Area),
	}

	game.Areas["кухня"] = &Area{
		Name:      "кухня",
		Neighbors: []string{"коридор"},
		Objects: map[string]string{
			"чай": "стол",
		},
	}
	game.Areas["коридор"] = &Area{
		Name:      "коридор",
		Neighbors: []string{"кухня", "комната", "улица"},
		Objects:   map[string]string{},
	}
	game.Areas["комната"] = &Area{
		Name:      "комната",
		Neighbors: []string{"коридор"},
		Objects: map[string]string{
			"ключи":     "стол",
			"конспекты": "стол",
			"рюкзак":    "стул",
		},
	}
	game.Areas["улица"] = &Area{
		Name:      "улица",
		Neighbors: []string{"домой"},
		Objects:   map[string]string{},
	}
}

func handleCommand(command string) string {
	parts := strings.Fields(command)
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
	area := game.Areas[game.Current]
	switch game.Current {
	case "кухня":
		places := describePlaces(area.Objects)
		suffix := "надо идти в универ."
		if !game.Player.BackPack {
			suffix = "надо собрать рюкзак и идти в универ."
		}
		return fmt.Sprintf(
			"ты находишься на кухне, %s, %s можно пройти - %s",
			places, suffix, strings.Join(area.Neighbors, ", "),
		)
	case "комната":
		if len(area.Objects) == 0 {
			return fmt.Sprintf(
				"пустая комната. можно пройти - %s",
				strings.Join(area.Neighbors, ", "),
			)
		}
		places := describePlaces(area.Objects)
		return fmt.Sprintf(
			"%s. можно пройти - %s",
			places, strings.Join(area.Neighbors, ", "),
		)
	default:
		return fmt.Sprintf(
			"ничего интересного. можно пройти - %s",
			strings.Join(area.Neighbors, ", "),
		)
	}
}

func cmdGo(dest string) string {
	area := game.Areas[game.Current]
	ok := false
	for _, n := range area.Neighbors {
		if n == dest {
			ok = true
			break
		}
	}
	if !ok {
		return "нет пути в " + dest
	}
	if dest == "улица" && !game.DoorOpen {
		return "дверь закрыта"
	}
	game.Current = dest
	switch dest {
	case "кухня":
		return fmt.Sprintf("кухня, ничего интересного. можно пройти - %s",
			strings.Join(game.Areas["кухня"].Neighbors, ", "))
	case "коридор":
		return fmt.Sprintf("ничего интересного. можно пройти - %s",
			strings.Join(game.Areas["коридор"].Neighbors, ", "))
	case "комната":
		return "ты в своей комнате. можно пройти - коридор"
	case "улица":
		return "на улице весна. можно пройти - домой"
	default:
		return fmt.Sprintf("ничего интересного. можно пройти - %s",
			strings.Join(game.Areas[dest].Neighbors, ", "))
	}
}

func cmdWear(item string) string {
	area := game.Areas[game.Current]
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
	area := game.Areas[game.Current]
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
	if target == "дверь" {
		if !game.DoorOpen {
			game.DoorOpen = true
			return "дверь открыта"
		}
		return "дверь уже открыта"
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

func main() {
}
