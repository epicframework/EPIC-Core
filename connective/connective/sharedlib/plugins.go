package main

import (
	"encoding/json"
	"errors"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"github.com/KernelDeimos/anything-gos/interp_a"
)

func makePluginFactory() interp_a.Operation {
	makeFunctions := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// The "device" plugin addes the add-device operation to a parent
	// operation. This makes it possible to add device entries which each have
	// their own thread-safe queue for property updates, as well as including
	// objects for meta information and Mozilla Web-of-Things definition.
	// See sub-commands of device plugin for more documentation.
	makeFunctions.AddOperation("device", makePlugDevice)

	return interp_a.Operation(makeFunctions.OpEvaluate)
}

type DevicePluginUpdateEvent map[string]interface{}

type DevicePluginEvent struct {
	Type     string                  `json:"type"`
	Contents DevicePluginUpdateEvent `json:"contents"`
}

type DeviceListEntry struct {
	InternalID    string             `json:"internal_id"`
	ManagerID     string             `json:"manager_id"`
	MozDefinition MozThingDefinition `json:"moz_definition"`
}

func makePlugDevice(args []interface{}) ([]interface{}, error) {
	//::gen verify-args makePlugDevice op interp_a.Operation
	if len(args) < 1 {
		return nil, errors.New("makePlugDevice requires at least 1 arguments")
	}

	var op interp_a.Operation
	{
		var ok bool
		op, ok = args[0].(interp_a.Operation)
		if !ok {
			return nil, errors.New("makePlugDevice: argument 0: op; must be type interp_a.Operation")
		}
	}
	//::end

	// var result []interface{}
	var err error

	// Add empty dirctory for device node registry
	deviceMapInternal := interp_a.InterpreterFactoryA{}.MakeEmpty()
	deviceMap := interp_a.InterpreterFactoryA{}.MakeEmpty()

	// Add an internal list of known devices (since listing in the directory
	// will display functions as well as devices)
	deviceList := []DeviceListEntry{}

	// Mutex lock for the device list
	mutexDeviceList := &sync.RWMutex{}

	//::run : testout (store (DATA))
	{
		r, e := op([]interface{}{"__debug_listmethods"})
		if e != nil {
			logrus.Error(e)
		}
		logrus.Debug(r)
	}
	//::end

	_, err = op([]interface{}{":", "internal_registry",
		interp_a.Operation(deviceMapInternal.OpEvaluate)})
	if err != nil {
		return nil, err
	}

	_, err = op([]interface{}{":", "registry",
		interp_a.Operation(deviceMap.OpEvaluate)})
	if err != nil {
		return nil, err
	}

	//::gen testout
	{
		r, e := op([]interface{}{"__debug_listmethods"})
		if e != nil {
			logrus.Error(e)
		}
		logrus.Debug(r)
	}
	//::end

	// Add list operation to devices operation
	_, err = op([]interface{}{":", "list", interp_a.Operation(func(
		args []interface{}) ([]interface{}, error) {
		mutexDeviceList.RLock()

		// Copy device list to []interface{} type
		result := []interface{}{}
		for _, item := range deviceList {
			result = append(result, item)
		}

		mutexDeviceList.RUnlock()

		return result, nil
	})})

	// Invoke "set" (:) operation to add add-device operation to operation
	// Usage: add-device <Mozilla definition> <user-defined meta information>
	_, err = op([]interface{}{":", "add-device", interp_a.Operation(func(
		args []interface{}) ([]interface{}, error) {
		//::gen verify-args add-device mozmeta interface{} usermeta interface{} externid string
		if len(args) < 3 {
			return nil, errors.New("add-device requires at least 3 arguments")
		}

		var mozmeta interface{}
		var usermeta interface{}
		var externid string
		{
			var ok bool
			mozmeta, ok = args[0].(interface{})
			if !ok {
				return nil, errors.New("add-device: argument 0: mozmeta; must be type interface{}")
			}
			usermeta, ok = args[1].(interface{})
			if !ok {
				return nil, errors.New("add-device: argument 1: usermeta; must be type interface{}")
			}
			externid, ok = args[2].(string)
			if !ok {
				return nil, errors.New("add-device: argument 2: externid; must be type string")
			}
		}
		//::end

		// Check if external ID is already in use
		{
			var index int
			var found bool
			for i := 0; i < len(deviceList); i++ {
				if deviceList[i].ManagerID == externid {
					found = true
					index = i
				}
			}

			if found {
				return []interface{}{deviceList[index].InternalID}, nil
			}
		}

		// Create UUID for the device
		deviceUUID := uuid.Must(uuid.NewV4()).String()
		logrus.Debug("Device id will be:", deviceUUID)

		// DECLARE struct for Mozilla web-thing definition
		var mozmetaStruct MozThingDefinition

		// DECLARE internal mutex for queue-properties interaction
		mutexProperties := &sync.RWMutex{}

		// Process Mozilla IoT definition
		{
			mozmetaBytes, err := json.Marshal(mozmeta)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(mozmetaBytes, &mozmetaStruct)
			if err != nil {
				return nil, err
			}
		}

		// Channels for cleanup messages
		stopPropertyUpdateQueue := make(chan struct{})
		// stopEventQueue := make(chan struct{})

		// Create device node object
		deviceNode := interp_a.InterpreterFactoryA{}.MakeEmpty()

		// --- device node operation to get Mozilla definition
		deviceNode.AddOperation("get-moz", func(
			args []interface{}) ([]interface{}, error) {
			return []interface{}{mozmetaStruct}, nil
		})

		// --- device node operation to get user-defined meta
		deviceNode.AddOperation("get-meta", func(
			args []interface{}) ([]interface{}, error) {
			return []interface{}{usermeta}, nil
		})

		// --- device node operation to close queues
		deviceNode.AddOperation("unlink", func(
			args []interface{}) ([]interface{}, error) {

			mutexDeviceList.RLock()

			var index int
			var found bool
			for i := 0; i < len(deviceList); i++ {
				if deviceList[i].InternalID == deviceUUID {
					index = i
					found = true
				}
			}

			mutexDeviceList.RUnlock()

			if !found {
				return nil, errors.New(
					"device list entry for `" + deviceUUID + "` not found")
			}

			mutexDeviceList.Lock()
			defer mutexDeviceList.Unlock()

			// Send signals to stop background operations
			stopPropertyUpdateQueue <- struct{}{}

			// Remove device from the device list
			deviceList = append(deviceList[:index], deviceList[index+1:]...)

			return []interface{}{usermeta}, nil
		})

		// Create a directory for properties
		{
			result, err := makeDSMap([]interface{}{})
			o := result[0].(interp_a.Operation)
			if err != nil {
				return nil, err
			}
			deviceNode.AddOperation("properties", o)
			// TODO: optimize properties by creating separate
			//       data-storage backend
		}

		// Create a directory for request action queues
		{

			actions := interp_a.InterpreterFactoryA{}.MakeEmpty()

			// Add each action to the actions map
			for actionName, _ := range mozmetaStruct.Actions {
				// Create a queue for device updatVes
				{
					result, err := makeDSQueue([]interface{}{})
					o := result[0].(interp_a.Operation)
					if err != nil {
						return nil, err
					}
					deviceNode.AddOperation("update-queue", o)

					// TODO: maybe it's possible to use Reflect in case an
					// uncast func(args...) is passed
					actionOp := result[0].(interp_a.Operation)

					actions.AddOperation(actionName, actionOp)
				}
			}

			// Add actions map to device node
			deviceNode.AddOperation("actions", actions.OpEvaluate)

		}

		// Create a queue for device updates
		{
			result, err := makeDSQueue([]interface{}{})
			o := result[0].(interp_a.Operation)
			if err != nil {
				return nil, err
			}
			deviceNode.AddOperation("update-queue", o)
		}

		// Create a queue for device property updates
		{
			result, err := makeDSBroadcastQueue([]interface{}{})
			o := result[0].(interp_a.Operation)
			if err != nil {
				return nil, err
			}
			deviceNode.AddOperation("event-queue", o)
		}

		// Start goroutine for queue-properties interaction
		go func() {
			// Custom error handler for this goroutine
			handleErrorInHere := func(err error) {
				// Log error to stderr
				logrus.Error(err)
				// Wait 20 seconds to prevent heavy log output
				<-time.After(20 * time.Second)
			}
			for {
				// Result goes in this channel after "update-queue block"
				resultFromUpdateQueue := make(chan []interface{})
				// Signal goes in this channel if "update-queue block" fails
				tryAgainFromUpdateQueue := make(chan struct{})

				// Perform blocking wait for update queue event
				go func() {
					result, err := deviceNode.OpEvaluate(
						[]interface{}{"update-queue", "block"})
					if err != nil {
						handleErrorInHere(err)
						tryAgainFromUpdateQueue <- struct{}{}
					}
					resultFromUpdateQueue <- result
				}()

				var result []interface{}

				select {
				case result = <-resultFromUpdateQueue:
				case <-tryAgainFromUpdateQueue:
					// Try "update-queue block" again
					continue
				case <-stopPropertyUpdateQueue:
					// Stop making requests to "update-queue block"
					return
				}

				if len(result) < 1 {
					logrus.Warnf("device '%s' received an empty event",
						deviceUUID,
					)
					continue
				}

				// Use intermediate JSON representation to normalize and
				// validate input event
				var event DevicePluginUpdateEvent
				{
					resultBytes, err := json.Marshal(result[0])
					if err != nil {
						handleErrorInHere(err)
						continue
					}

					err = json.Unmarshal(resultBytes, &event)
					if err != nil {
						handleErrorInHere(err)
						continue
					}
				}

				// Send the update to the event queue
				_, err = deviceNode.OpEvaluate(
					[]interface{}{"event-queue", "enque", DevicePluginEvent{
						Type:     "property.set",
						Contents: event,
					}})
				if err != nil {
					handleErrorInHere(err)
				}

				// Perform the update on internal storage
				mutexProperties.Lock()
				for key, val := range event {
					logrus.Debugf("Updating %s/%s with %v",
						deviceUUID, key, val,
					)
					// Update property using set (:) operation
					result, err := deviceNode.OpEvaluate([]interface{}{
						"properties", ":", key, interp_a.Operation(
							// Use anonymous function to wrap data;
							// this is a temporary solution until the hashmap
							// storage backend is implemented in interp_a
							func(_ []interface{}) ([]interface{}, error) {
								return []interface{}{val}, nil
							},
						)})
					if err != nil {
						logrus.Error(err)
						logrus.Debug(result)
						continue
					}
				}

				mutexProperties.Unlock()

			}
		}()

		// Register device to map of internal UUIDs
		deviceMapInternal.AddOperation(deviceUUID, deviceNode.OpEvaluate)
		// Register device to map of external UUIDs
		deviceMap.AddOperation(externid, deviceNode.OpEvaluate)

		// Add device to the device list
		mutexDeviceList.Lock()
		deviceList = append(deviceList, DeviceListEntry{
			InternalID:    deviceUUID,
			ManagerID:     externid,
			MozDefinition: mozmetaStruct,
		})
		mutexDeviceList.Unlock()

		// Report HA/Connective's internal UUID
		return []interface{}{deviceUUID}, nil
	})}) // geez this is starting to look like Javascript
	if err != nil {
		return nil, err
	}

	/*
		// Invoke "set" (:) operation to add add-device operation to operation
		// Usage: add-device <Mozilla definition> <user-defined meta information>
		_, err = op([]interface{}{":", "unlink-device", interp_a.Operation(func(
			args []interface{}) ([]interface{}, error) {
			//::gen verify-args unlink-device mozmeta interface{} usermeta interface{} externid string
			if len(args) < 3 {
				return nil, errors.New("unlink-device requires at least 3 arguments")
			}

			var mozmeta interface{}
			var usermeta interface{}
			var externid string
			{
				var ok bool
				mozmeta, ok = args[0].(interface{})
				if !ok {
					return nil, errors.New("unlink-device: argument 0: mozmeta; must be type interface{}")
				}
				usermeta, ok = args[1].(interface{})
				if !ok {
					return nil, errors.New("unlink-device: argument 1: usermeta; must be type interface{}")
				}
				externid, ok = args[2].(string)
				if !ok {
					return nil, errors.New("unlink-device: argument 2: externid; must be type string")
				}
			}
			//::end
			})}
		if err != nil {
			return nil, err
		}
	*/

	return nil, nil
}
