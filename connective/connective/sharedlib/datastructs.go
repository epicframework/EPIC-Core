package main

import (
	"crypto/sha1"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"github.com/KernelDeimos/anything-gos/interp_a"
	"github.com/KernelDeimos/gottagofast/utiltime"
)

/*
	This file contains data strictures which can by used via the interpreter.
*/

func makeDataStructureFactory() interp_a.Operation {
	makeFunctions := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// A "directory" is internally referred to as a map. It is just an empty
	// interpreter instance on which other data structures can be applied with
	// a name. For example, "sensors" will be a directory containing request
	// queues.
	makeFunctions.AddOperation("directory", makeDSMap)

	// Making the type "requests" will create an interpreter instance with
	// the following functions:
	// - flush   -> report a list of requests and clear queue
	// - block   -> report the oldest request with blocking wait
	// - check   -> report the oldest request or nil
	// - enque   -> add some data to the request queue (blocking if full queue)
	makeFunctions.AddOperation("requests", makeDSQueue)

	makeFunctions.AddOperation("heartbeat-monitor", makeDSHeartbeatMonitor)

	return interp_a.Operation(makeFunctions.OpEvaluate)
}

func makeDSMap(args []interface{}) ([]interface{}, error) {
	empty := interp_a.InterpreterFactoryA{}.MakeEmpty()
	return []interface{}{interp_a.Operation(empty.OpEvaluate)}, nil
}

func makeDSQueue(args []interface{}) ([]interface{}, error) {
	size := 100
	if len(args) > 0 {
		var ok bool
		size, ok = args[0].(int)
		if !ok {
			return nil, errors.New("request queue size must be integer")
		}
	}

	// Create empty interpreter for request queue functions
	empty := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// Make request queue
	ch := make(chan interface{}, size)
	queue := RequestQueue{
		Chan: ch,
	}

	// Bind request queue functions to interpreter
	queue.Bind(empty)

	return []interface{}{interp_a.Operation(empty.OpEvaluate)}, nil
}

func makeDSBroadcastQueue(args []interface{}) ([]interface{}, error) {
	// Create empty interpreter for request queue functions
	empty := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// Make request queue
	ch := make(chan interface{})
	queue := BroadcastQueue{
		Chan: ch,
	}

	// Bind request queue functions to interpreter
	queue.Bind(empty)

	return []interface{}{interp_a.Operation(empty.OpEvaluate)}, nil
}

func makeDSHeartbeatMonitor(args []interface{}) ([]interface{}, error) {
	//::gen verify-args make-heartbeat-monitor timeoutstr string
	if len(args) < 1 {
		return nil, errors.New("make-heartbeat-monitor requires at least 1 arguments")
	}

	var timeoutstr string
	{
		var ok bool
		timeoutstr, ok = args[0].(string)
		if !ok {
			return nil, errors.New("make-heartbeat-monitor: argument 0: timeoutstr; must be type string")
		}
	}
	//::end

	timeout, err := time.ParseDuration(timeoutstr)
	if err != nil {
		return nil, err
	}

	// Create empty interpreter for heartbeat functions
	empty := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// Create heartbeat monitor data
	hm := &HeartbeatMonitor{
		LastBeat: time.Now(),
		Timeout:  timeout,
		Mutex:    &sync.RWMutex{},
	}

	// Bind heartbeat functions to interpreter
	hm.Bind(empty)
	return []interface{}{interp_a.Operation(empty.OpEvaluate)}, nil
}

type RequestQueue struct {
	Chan chan interface{}
}

func (rq RequestQueue) Bind(destination interp_a.HybridEvaluator) {
	destination.AddOperation("enque", rq.OpEnqueue)
	destination.AddOperation("block", rq.OpDequeueBlk)
}

func (rq RequestQueue) OpEnqueue(args []interface{}) ([]interface{}, error) {
	for _, arg := range args {
		rq.Chan <- arg
	}
	return nil, nil
}

func (rq RequestQueue) OpDequeueBlk(args []interface{}) ([]interface{}, error) {
	value := <-rq.Chan
	return []interface{}{value}, nil
}

type BroadcastQueue struct {
	Chan chan interface{}
}

func (rq BroadcastQueue) Bind(destination interp_a.HybridEvaluator) {
	destination.AddOperation("enque", rq.OpEnqueue)
	destination.AddOperation("block", rq.OpDequeueBlk)
}

func (rq BroadcastQueue) OpEnqueue(args []interface{}) ([]interface{}, error) {
	sent := 0
	for _, arg := range args {
		select {
		case rq.Chan <- arg:
			sent++
		default:
			//
		}
	}
	return []interface{}{sent}, nil
}

func (rq BroadcastQueue) OpDequeueBlk(args []interface{}) ([]interface{}, error) {
	value := <-rq.Chan
	return []interface{}{value}, nil
}

type HashGraph struct {
	// All nodes of graph, including "meta nodes" for links
	Nodes map[string]interface{}

	// List of nodes that are not meta nodes for links
	VisibleNodes map[string]struct{}

	// First key is a node on one side of the link
	// (so a link will have two entries)
	// Second key is the "meta node" for the link
	Links map[string]map[string]struct{}
}

func (hg *HashGraph) AddLink(nodeKey, linkKey string) {
	links, exists := hg.Links[nodeKey]
	if !exists {
		links = map[string]struct{}{}
	}
	links[linkKey] = struct{}{}
	hg.Links[nodeKey] = links
}

func (hg *HashGraph) OpAdd(args []interface{}) ([]interface{}, error) {
	for _, arg := range args {
		// Determine key
		var key string
		switch a := arg.(type) {
		case string:
			k := sha1.Sum([]byte(a))
			key = string(k[:])
		case []byte:
			k := sha1.Sum(a)
			key = string(k[:])
		default:
			uuid, err := uuid.NewV4()
			if err != nil {
				return []interface{}{}, err
			}
			key = uuid.String()
		}

		hg.Nodes[key] = arg
		hg.VisibleNodes[key] = struct{}{}
	}
	return []interface{}{}, nil
}

func (hg *HashGraph) OpLink(args []interface{}) ([]interface{}, error) {
	// Verify parameters (code below is generated using genfor-interp-a)
	//::gen verify-args link keyA string keyB string value interface{}
	if len(args) < 3 {
		return nil, errors.New("link requires at least 3 arguments")
	}

	var keyA string
	var keyB string
	var value interface{}
	{
		var ok bool
		keyA, ok = args[0].(string)
		if !ok {
			return nil, errors.New("link: argument 0: keyA; must be type string")
		}
		keyB, ok = args[1].(string)
		if !ok {
			return nil, errors.New("link: argument 1: keyB; must be type string")
		}
		value, ok = args[2].(interface{})
		if !ok {
			return nil, errors.New("link: argument 2: value; must be type interface{}")
		}
	}
	//::end

	// Sort keys
	keys := []string{keyA, keyB}
	sort.Strings(keys)
	keyA = keys[0]
	keyB = keys[1]

	hashBytes := sha1.Sum([]byte(keyA + "." + keyB))
	hash := string(hashBytes[:])

	// Add link
	hg.Nodes[hash] = value
	hg.AddLink(keyA, hash)
	hg.AddLink(keyB, hash)

	return []interface{}{}, nil
}

func (hg *HashGraph) OpGetNodes(args []interface{}) ([]interface{}, error) {
	ii := []interface{}{}
	for _, key := range hg.VisibleNodes {
		ii = append(ii, key)
	}
	return ii, nil
}

func (hg *HashGraph) OpGet(args []interface{}) ([]interface{}, error) {
	//::gen verify-args get key string
	if len(args) < 1 {
		return nil, errors.New("get requires at least 1 arguments")
	}

	var key string
	{
		var ok bool
		key, ok = args[0].(string)
		if !ok {
			return nil, errors.New("get: argument 0: key; must be type string")
		}
	}
	//::end
	ii := []interface{}{}
	ii = append(ii, hg.Nodes[key])
	return ii, nil
}

func (hg *HashGraph) OpGetLinks(args []interface{}) ([]interface{}, error) {
	//::gen verify-args get-links key string
	if len(args) < 1 {
		return nil, errors.New("get-links requires at least 1 arguments")
	}

	var key string
	{
		var ok bool
		key, ok = args[0].(string)
		if !ok {
			return nil, errors.New("get-links: argument 0: key; must be type string")
		}
	}
	//::end

	links, _ := hg.Links[key]

	ii := []interface{}{}
	for link := range links {
		ii = append(ii, link)
	}

	return ii, nil
}

type HeartbeatMonitor struct {
	// Last time a heartbeat was recieved
	LastBeat time.Time

	// Callback to trigger when a process appears to be dead
	DeadFunc interp_a.Operation

	// Amount of time before DeadFunc is triggered after the last heartbeat
	Timeout time.Duration

	// Ticker for the heartbeat monitor, or nil if the monitor is inactive.
	// This is exposed to functions outside the monitor because any heartbeat
	// update should reset the timer to ensure prompt detection of inactivity.
	Ticker *utiltime.RealTicker

	// RWMutex for heartbeat updates
	Mutex *sync.RWMutex
}

func (hm *HeartbeatMonitor) Bind(destination interp_a.HybridEvaluator) {
	destination.AddOperation("monitor", hm.OpMonitor)
	destination.AddOperation("beat", hm.OpBeat)
	destination.AddOperation("is-alive", hm.OpIsAlive)
	destination.AddOperation("time-since", hm.OpTimeSince)
}

// OpMonitor starts a heartbeat monitor implementation
func (hm *HeartbeatMonitor) OpMonitor(args []interface{}) ([]interface{}, error) {
	//::gen verify-args heartbeat-monitor callback interp_a.Operation
	if len(args) < 1 {
		return nil, errors.New("heartbeat-monitor requires at least 1 arguments")
	}

	var callback interp_a.Operation
	{
		var ok bool
		callback, ok = args[0].(interp_a.Operation)
		if !ok {
			return nil, errors.New("heartbeat-monitor: argument 0: callback; must be type interp_a.Operation")
		}
	}
	//::end

	// Set callback for heartbeat timeout
	hm.DeadFunc = callback

	// Start routine to periodically check heartbeat
	go func() {
		hm.Ticker = utiltime.NewRealTicker(hm.Timeout)
		for {
			// Wait the duration of one timeout
			<-hm.Ticker.C

			// Get last beat
			hm.Mutex.RLock()
			t := hm.LastBeat
			hm.Mutex.RUnlock()

			duration := time.Now().Sub(t)
			if duration >= hm.Timeout {
				hm.DeadFunc([]interface{}{})
			}
		}
	}()

	return []interface{}{}, nil
}

func (hm *HeartbeatMonitor) OpBeat(args []interface{}) ([]interface{}, error) {
	hm.Mutex.Lock()
	defer hm.Mutex.Unlock()
	hm.LastBeat = time.Now()
	if hm.Ticker != nil {
		hm.Ticker.Reset()
	}
	return []interface{}{}, nil
}

func (hm *HeartbeatMonitor) OpIsAlive(args []interface{}) ([]interface{}, error) {
	hm.Mutex.RLock()
	t := hm.LastBeat
	hm.Mutex.RUnlock()

	duration := time.Now().Sub(t)
	if duration >= hm.Timeout {
		return []interface{}{false}, nil
	}

	return []interface{}{true}, nil
}

func (hm *HeartbeatMonitor) OpTimeSince(args []interface{}) ([]interface{}, error) {
	hm.Mutex.RLock()
	t := hm.LastBeat
	hm.Mutex.RUnlock()

	duration := time.Now().Sub(t)

	return []interface{}{int64(duration / time.Second)}, nil
}
