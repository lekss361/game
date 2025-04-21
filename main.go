package main

import (
	"fmt"
	"sort"
	"strings"
)

type ActionFunc func(g *Game, args ...string) string

type Person struct {
	BackPack  bool
	Inventory map[string]int
}

type Game struct {
	Current  *Area
	Areas    map[string]*Area
	Player   Person
	DoorOpen bool
	Commands map[string]ActionFunc
}

type Area struct {
	Name           string
	Description    string
	Neighbors      map[string]*Area
	NeighborsOrder []string
	Objects        map[string]string
	UseActions     map[string]map[string]func(*Game) string
	LookAction     ActionFunc
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
	kitchen.NeighborsOrder = []string{"коридор"}

	corridor.Neighbors = map[string]*Area{
		"кухня":   kitchen,
		"комната": room,
		"улица":   street,
	}
	corridor.NeighborsOrder = []string{"кухня", "комната", "улица"}

	room.Neighbors = map[string]*Area{"коридор": corridor}
	room.NeighborsOrder = []string{"коридор"}

	street.Neighbors = map[string]*Area{"домой": corridor}
	street.NeighborsOrder = []string{"домой"}

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

	defaultLook := func(g *Game, _ ...string) string {
		return fmt.Sprintf(
			"ничего интересного. можно пройти - %s",
			neighborsList(g.Current),
		)
	}
	kitchen.LookAction = func(g *Game, _ ...string) string {
		places := describePlaces(g.Current.Objects)
		suffix := "надо идти в универ."
		if !g.Player.BackPack {
			suffix = "надо собрать рюкзак и идти в универ."
		}
		return fmt.Sprintf(
			"ты находишься на кухне, %s, %s можно пройти - %s",
			places, suffix, neighborsList(g.Current),
		)
	}
	room.LookAction = func(g *Game, _ ...string) string {
		if len(g.Current.Objects) == 0 {
			return fmt.Sprintf(
				"пустая комната. можно пройти - %s",
				neighborsList(g.Current),
			)
		}
		places := describePlaces(g.Current.Objects)
		return fmt.Sprintf(
			"%s. можно пройти - %s",
			places, neighborsList(g.Current),
		)
	}
	corridor.LookAction = defaultLook
	street.LookAction = defaultLook

	commands := map[string]ActionFunc{
		"осмотреться": func(g *Game, _ ...string) string { return g.Current.LookAction(g) },
		"идти":        func(g *Game, args ...string) string { return g.Current.Go(g, args) },
		"надеть":      func(g *Game, args ...string) string { return g.Current.Wear(g, args) },
		"взять":       func(g *Game, args ...string) string { return g.Current.Take(g, args) },
		"применить":   func(g *Game, args ...string) string { return g.Current.Use(g, args) },
	}

	game = Game{
		Current:  kitchen,
		Areas:    map[string]*Area{"кухня": kitchen, "коридор": corridor, "комната": room, "улица": street},
		Player:   Person{BackPack: false, Inventory: make(map[string]int)},
		DoorOpen: false,
		Commands: commands,
	}
}

func handleCommand(cmd string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "неизвестная команда"
	}
	action, ok := game.Commands[parts[0]]
	if !ok {
		return "неизвестная команда"
	}
	return action(&game, parts[1:]...)
}

func (a *Area) Go(g *Game, args []string) string {
	if len(args) < 1 {
		return "неизвестная команда"
	}
	dest := args[0]
	next, ok := a.Neighbors[dest]
	if !ok {
		return "нет пути в " + dest
	}
	if dest == "улица" && !g.DoorOpen {
		return "дверь закрыта"
	}
	g.Current = next
	return fmt.Sprintf(
		"%s. можно пройти - %s",
		next.Description,
		neighborsList(next),
	)
}

func (a *Area) Wear(g *Game, args []string) string {
	if len(args) < 1 {
		return "неизвестная команда"
	}
	item := args[0]
	if item == "рюкзак" {
		if _, ok := a.Objects["рюкзак"]; ok {
			g.Player.BackPack = true
			delete(a.Objects, "рюкзак")
			return "вы надели: рюкзак"
		}
	}
	return "нет такого"
}

func (a *Area) Take(g *Game, args []string) string {
	if len(args) < 1 {
		return "неизвестная команда"
	}
	item := args[0]
	if !g.Player.BackPack {
		return "некуда класть"
	}
	if _, ok := a.Objects[item]; !ok {
		return "нет такого"
	}
	g.Player.Inventory[item]++
	delete(a.Objects, item)
	return "предмет добавлен в инвентарь: " + item
}

func (a *Area) Use(g *Game, args []string) string {
	if len(args) < 2 {
		return "неизвестная команда"
	}
	item, target := args[0], args[1]
	if g.Player.Inventory[item] == 0 {
		return "нет предмета в инвентаре - " + item
	}
	if acts, ok := a.UseActions[item]; ok {
		if action, ok2 := acts[target]; ok2 {
			return action(g)
		}
	}
	return "не к чему применить"
}

func describePlaces(objects map[string]string) string {
	loc := map[string]string{"стол": "столе", "стул": "стуле"}
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
	return strings.Join(area.NeighborsOrder, ", ")
}

func main() {
}
