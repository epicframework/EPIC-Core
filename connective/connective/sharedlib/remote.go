package main

import (
	"C"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/KernelDeimos/anything-gos/interp_a"
)

type Receipt struct {
	Source    string
	Event     string
	Command   string
	TimeStamp int64 `json:"Time Stamp"`
	Sequence  int64
	Result    string
}

type ServerReceipt struct {
	Dispatched bool `json:"dispatched"`
	Completed  bool `json:"completed"`
}

//export elconn_serve_remote
func elconn_serve_remote(addr *C.char, opID LibSharedID) int32 {
	// Obtain caller inputs
	opInterface, okay := GetSharedItem(LibSharedTypeAPI, opID)
	if !okay {
		logrus.Error("call operation: invalid value")
		return 0
	}
	addrStr := C.GoString(addr)

	// Dereference caller inputs
	op := (*opInterface).(interp_a.Operation)

	// Add middleware for meta information
	var serverOp interp_a.Operation
	{
		var cmdGetReceipt interp_a.Operation
		serverOp, cmdGetReceipt = newReceiptStore(op)

		// Execute command for getting the receipt
		serverOp([]interface{}{":", "get-receipt", cmdGetReceipt})
	}

	router := gin.Default()
	router.POST("/call", func(c *gin.Context) {
		listStr := c.PostForm("list")
		list := []interface{}{}

		err := json.Unmarshal([]byte(listStr), &list)
		if err != nil {
			logrus.Error(err)
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		result, err := serverOp(list)
		if err != nil {
			logrus.Error(err)
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.JSON(http.StatusOK, result)

	})

	go func() {
		err := router.Run(addrStr)
		if err != nil {
			logrus.Error(err)
		}
	}()

	return 0
}

//export elconn_connect_remote
func elconn_connect_remote(addr *C.char) LibSharedID {
	// Obtain caller inputs
	addrStr := C.GoString(addr) + "/call"

	// Create HTTP client
	client := http.Client{}

	// Define operation to send request
	op := func(args []interface{}) ([]interface{}, error) {
		list, err := json.Marshal(args)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		// Create form data
		form := url.Values{}
		form.Add("list", string(list))

		// Create request object (and encode form to do so)
		req, err := http.NewRequest("POST", addrStr, strings.NewReader(
			form.Encode(),
		))
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Perform request
		resp, err := client.Do(req)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		// Read response
		// -- read response as bytes
		defer resp.Body.Close()
		responseBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		// -- parse response bytes as JSON
		responseList := []interface{}{}
		err = json.Unmarshal(responseBytes, &responseList)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		return responseList, nil
	}

	// Generate UUID to identify this source
	uuidStr, err := uuid.NewRandom()
	if err != nil {
		return AddSharedError(err)
	}

	// Add a middleware that adds meta to requests and prints receipts
	var opInterface interface{}
	opInterface = newReceiptPrinter(
		uuidStr.String(),
		interp_a.Operation(op),
	)

	id := AddSharedItem(LibSharedTypeAPI, &opInterface)
	return id
}

func newReceiptPrinter(
	source string, dest interp_a.Operation,
) interp_a.Operation {
	// TODO: mutex lock for seqNum
	seqNumMutex := &sync.Mutex{}
	var seqNum int

	return func(args []interface{}) ([]interface{}, error) {
		// Create command id:
		var commandID string
		{
			// Hash the command instructions
			// (
			//  not strictly necessary for identification,
			//  but useful to verify the integrity of the command.
			// )
			data, err := json.Marshal(args)
			if err != nil {
				return nil, err
			}
			hash := sha1.Sum(data)
			hashStr := hex.EncodeToString(hash[:])

			// Generate ID using source UUID + sequence number + hash
			seqNumMutex.Lock()
			seqNum++
			commandID = source + "." + strconv.Itoa(seqNum) + "." + hashStr
			seqNumMutex.Unlock()
		}
		eventStr := "Unknown"
		if len(args) > 1 {
			eventStr = fmt.Sprint(args[0])
		}
		receipt, err := json.Marshal(Receipt{
			Event:   eventStr,
			Command: commandID,
		})
		if err != nil {
			return nil, err
		}
		fmt.Println(string(receipt))

		// Prepend meta information to command arguments (for server)
		{
			metaArgs := []interface{}{"__META__"}
			metaArgs = append(metaArgs, "cmdid="+commandID)
			metaArgs = append(metaArgs, "__ENDMETA__")

			args = append(metaArgs, args...)
		}

		return dest(args)
	}
}

func newReceiptStore(dest interp_a.Operation) (
	interp_a.Operation, // Middleware entry point
	interp_a.Operation, // Receipt get command
) {
	mutex := &sync.Mutex{}
	receiptStore := map[string]ServerReceipt{}

	// Middleware operation to parse meta information inside commands.
	// Meta goes into a section beginning with __META__ and ending with
	// __ENDMETA__
	//
	// Example:
	//   original command:
	//     heartbeats a beat
	//   cmd with meta:
	//     __META__ source=a cmdid=... __ENDMETA__ heartbeats a beat
	op := func(args []interface{}) ([]interface{}, error) {

		// This block goes to DoNormalCommand immediately if the
		// arguments do not begin with __META__
		{
			if len(args) < 1 {
				goto DoNormalCommand
			}
			if val, ok := args[0].(string); ok {
				if val != "__META__" {
					goto DoNormalCommand
				}
			} else {
				goto DoNormalCommand
			}
		}

		// Scope limiter for conditional local memory
		{
			// Copy all key-value pairs until __ENDMETA__
			metaValues := map[string]string{}
			for i := 1; i < len(args); i++ {
				value, ok := args[i].(string)
				if !ok {
					return nil, errors.New(
						"Remote operation meta must be all strings",
					)
				}

				// Check for __ENDMETA__
				if value == "__ENDMETA__" {
					// Store receipt verification entry
					if cmdid, exists := metaValues["cmdid"]; exists {
						func() {
							mutex.Lock()
							defer mutex.Unlock()
							receiptStore[cmdid] = ServerReceipt{
								Dispatched: true,
							}
						}()

						// Deferred function is executed after the block at
						// "DoNormalCommand" label completes
						defer func() {
							mutex.Lock()
							defer mutex.Unlock()
							entry := receiptStore[cmdid]
							entry.Completed = true
							receiptStore[cmdid] = entry
						}()
					}

					// Run intended command
					args = args[i+1:]
					goto DoNormalCommand
				}

				// Parse the next key-value pair in this meta section
				valueParts := strings.Split(value, "=")
				if len(valueParts) != 2 {
					return nil, errors.New(
						"All meta strings must be key=value pairs",
					)
				}
				metaValues[valueParts[0]] = strings.Join(valueParts[1:], "=")
			}
		}

	DoNormalCommand:
		return dest(args)
	}

	receiptGetCmd := func(args []interface{}) ([]interface{}, error) {
		//::gen verify-args receipt-get commandid string
		if len(args) < 1 {
			return nil, errors.New("receipt-get requires at least 1 arguments")
		}

		var commandid string
		{
			var ok bool
			commandid, ok = args[0].(string)
			if !ok {
				return nil, errors.New("receipt-get: argument 0: commandid; must be type string")
			}
		}
		//::end

		mutex.Lock()
		entry, exists := receiptStore[commandid]
		mutex.Unlock()

		if !exists {
			return []interface{}{"unrecognized"}, nil
		}

		switch true {
		case entry.Completed && entry.Dispatched:
			return []interface{}{"completed"}, nil
		case entry.Dispatched:
			return []interface{}{"started"}, nil
		}

		return []interface{}{"inconsistent"}, nil
	}

	return op, receiptGetCmd
}
