package engine

import (
	"fmt"
	"github.com/ivorytoast/replay78/states"
	"log"
	"os"
	"strings"
	"time"
)

type Application interface {
	OnEvent(event []string)
	Topics() []string
}

type InputGenerator interface {
	Start(engine *Engine)
}

type IntervalGenerator struct {
	InputFunc func() string
	Interval  time.Duration
}

type ConnectionGenerator struct {
	StartFunc func(engine *Engine)
}

func (g *IntervalGenerator) Start(e *Engine) {
	go func() {
		for {
			input := g.InputFunc()
			e.In(input)
			time.Sleep(g.Interval)
		}
	}()
	log.Printf("Started interval generator (interval: %v)", g.Interval)
}

func (g *ConnectionGenerator) Start(e *Engine) {
	go func() {
		log.Printf("Started connection generator")
		g.StartFunc(e)
	}()
}

type Engine struct {
	file         *os.File
	seq          int
	queue        chan string
	applications map[string]Application
	generators   []InputGenerator

	TicTacToeState *states.TicTacToeState
}

func NewEngine() *Engine {
	// Find the highest numbered log file
	highestNum := 0
	for i := 1; i < 1000; i++ {
		filename := fmt.Sprintf("78-%d.log", i)
		if _, err := os.Stat(filename); err == nil {
			highestNum = i
		} else {
			break
		}
	}

	// Create the next log file
	nextLogFile := fmt.Sprintf("78-%d.log", highestNum+1)
	return NewEngineWithLogFile(nextLogFile)
}

func NewEngineWithLogFile(logFileName string) *Engine {
	if _, err := os.Stat(logFileName); err == nil {
		// Extract base name and current number
		base := strings.TrimSuffix(logFileName, ".log")
		startNum := 1

		// Check if filename already has a number (e.g., "78-1.log")
		parts := strings.Split(base, "-")
		if len(parts) > 1 {
			// Try to parse the last part as a number
			var num int
			if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &num); err == nil {
				startNum = num + 1
				base = strings.Join(parts[:len(parts)-1], "-")
			}
		}

		// Find next available number
		for i := startNum; ; i++ {
			rotated := fmt.Sprintf("%s-%d.log", base, i)
			if _, err := os.Stat(rotated); os.IsNotExist(err) {
				os.Rename(logFileName, rotated)
				break
			}
		}
	}
	f, _ := os.Create(logFileName)
	return &Engine{
		file:           f,
		queue:          make(chan string, 100),
		applications:   make(map[string]Application),
		generators:     make([]InputGenerator, 0),
		seq:            0,
		TicTacToeState: states.NewTicTacToeState(),
	}
}

func (e *Engine) RegisterApplication(app Application) {
	for _, topic := range app.Topics() {
		e.applications[topic] = app
	}
}

func (e *Engine) RegisterGenerator(gen InputGenerator) {
	e.generators = append(e.generators, gen)
}

func (e *Engine) TTT() *states.TicTacToeState {
	return e.TicTacToeState
}

func (e *Engine) nextSeq() int {
	e.seq++
	return e.seq
}

func (e *Engine) Run() {
	go e.run()
}

func (e *Engine) run() {
	for _, gen := range e.generators {
		gen.Start(e)
	}
	for line := range e.queue {
		parts, isValid := parseMsg(line)
		if !isValid {
			e.Out("Bad Input: " + line)
			continue
		}
		topic := parts[0]
		action := parts[1]
		payload := parts[2]
		e.file.WriteString(fmt.Sprintf("%d|I|%s|%s|%s\n", e.nextSeq(), topic, action, payload))
		if app, ok := e.applications[topic]; ok {
			app.OnEvent(parts)
		}
	}
}

func (e *Engine) In(line string) {
	e.queue <- line
}

func (e *Engine) Out(line string) {
	e.file.WriteString(fmt.Sprintf("%d|O|%s\n", e.nextSeq(), line))
}

func parseMsg(line string) ([]string, bool) {
	parts := strings.SplitN(line, "|", 3)
	if len(parts) < 3 {
		return parts, false
	}
	return parts, true
}
