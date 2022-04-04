package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mattn/go-tty"
)

type game struct {
	score int
	snake *snake
	food  position
}

type snake struct {
	body      []position
	direction direction
}

type position [2]int

type direction int

const (
	north direction = iota
	east
	south
	west
)

func main() {
	game := newGame()
	game.beforeGame()

	for {
		maxX, maxY := getSize()

		// calculate new head position
		newHeadPos := game.snake.body[0]

		switch game.snake.direction {
		case north:
			newHeadPos[1]--
		case east:
			newHeadPos[0]++
		case south:
			newHeadPos[1]++
		case west:
			newHeadPos[0]--
		}

		// if you hit the wall, game over
		hitWall := newHeadPos[0] < 1 || newHeadPos[1] < 1 || newHeadPos[0] > maxX ||
			newHeadPos[1] > maxY
		if hitWall {
			game.over()
		}

		// if you run into yourself, game over
		for _, pos := range game.snake.body {
			if positionsAreSame(newHeadPos, pos) {
				game.over()
			}
		}

		// add the new head to the body
		game.snake.body = append([]position{newHeadPos}, game.snake.body...)

		ateFood := positionsAreSame(game.food, newHeadPos)
		if ateFood {
			game.score++
			game.placeNewFood()
		} else {
			game.snake.body = game.snake.body[:len(game.snake.body)-1]
		}

		game.draw()
	}
}

func newGame() *game {
	rand.Seed(time.Now().UnixNano())

	snake := newSnake()

	game := &game{
		score: 0,
		snake: snake,
		food:  randomPosition(),
	}

	go game.listenForKeyPress()

	return game
}

func positionsAreSame(a, b position) bool {
	return a[0] == b[0] && a[1] == b[1]
}

func randomPosition() position {
	width, height := getSize()
	x := rand.Intn(width) + 1
	y := rand.Intn(height) + 2

	return [2]int{x, y}
}

func newSnake() *snake {
	maxX, maxY := getSize()
	pos := position{maxX / 2, maxY / 2}

	return &snake{
		body:      []position{pos},
		direction: north,
	}
}

func (g *game) draw() {
	clear()
	maxX, _ := getSize()

	status := "score: " + strconv.Itoa(g.score)
	statusXPos := maxX/2 - len(status)/2

	moveCursor(position{statusXPos, 0})
	draw(status)

	moveCursor(g.food)
	draw("*")

	for i, pos := range g.snake.body {
		moveCursor(pos)

		if i == 0 {
			draw("O")
		} else {
			draw("o")
		}
	}

	render()
	time.Sleep(time.Millisecond * 50)
}

func (g *game) beforeGame() {
	hideCursor()

	// handle CTRL C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			g.over()
		}
	}()
}

func (g *game) over() {
	clear()
	showCursor()

	moveCursor(position{1, 1})
	draw("game over. score: " + strconv.Itoa(g.score))

	render()

	os.Exit(0)
}

func (g *game) placeNewFood() {
	for {
		newFoodPosition := randomPosition()

		if positionsAreSame(newFoodPosition, g.food) {
			continue
		}

		for _, pos := range g.snake.body {
			if positionsAreSame(newFoodPosition, pos) {
				continue
			}
		}

		g.food = newFoodPosition

		break
	}
}

func (g *game) listenForKeyPress() {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		char, err := tty.ReadRune()
		if err != nil {
			panic(err)
		}

		// UP, DOWN, RIGHT, LEFT == [A, [B, [C, [D
		// we ignore the escape character [
		switch char {
		case 'A':
			g.snake.direction = north
		case 'B':
			g.snake.direction = south
		case 'C':
			g.snake.direction = east
		case 'D':
			g.snake.direction = west
		}
	}
}
