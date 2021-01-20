package tables

import (
	"regexp"
	"sync"

	"github.com/Zamiell/hanabi-live/server/pkg/dispatcher"
	"github.com/Zamiell/hanabi-live/server/pkg/logger"
	"github.com/Zamiell/hanabi-live/server/pkg/models"
	"github.com/Zamiell/hanabi-live/server/pkg/table"
	"github.com/tevino/abool"
)

// Manager is an object that handles adding or removing tables,
// as well as user relationships to tables.
// It listens for requests in a new goroutine.
type Manager struct {
	name string

	tables           map[int]*table.Manager // Indexed by table ID
	tableIDCounter   int
	usersPlaying     map[int][]int // Indexed by user ID, values are table IDs
	usersSpectating  map[int][]int // Indexed by user ID, values are table IDs
	isValidTableName func(string) bool

	requests          chan *request
	requestsWaitGroup sync.WaitGroup
	requestFuncMap    map[requestType]func(interface{}) interface{}
	requestsClosed    *abool.AtomicBool

	logger     *logger.Logger
	models     *models.Models
	Dispatcher *dispatcher.Dispatcher

	useTLS bool
	domain string
}

func NewManager(logger *logger.Logger, models *models.Models, useTLS bool, domain string) *Manager {
	m := &Manager{
		name: "tables",

		tables:           make(map[int]*table.Manager),
		tableIDCounter:   0, // The first table ID will be 1 and will increase from there
		usersPlaying:     make(map[int][]int),
		usersSpectating:  make(map[int][]int),
		isValidTableName: regexp.MustCompile(`^[a-zA-Z0-9 !@#$\(\)\-_=\+;:,\.\?]+$`).MatchString,

		requests:       make(chan *request),
		requestFuncMap: make(map[requestType]func(interface{}) interface{}),
		requestsClosed: abool.New(),

		logger:     logger,
		models:     models,
		Dispatcher: nil, // This will be filled in after this object is instantiated

		useTLS: useTLS,
		domain: domain,
	}
	m.requestFuncMapInit()
	go m.ListenForRequests()

	return m
}