# HA/Connective

This is the connective component for the hub application.

The connective component creates an abstract model for IoT devices, and in
particular the sensor-actuator nodes for HVAC control. This component is
implemented as a shared library. (.so for Linux target, .dll in Windows)
This allows interoperability between components written in different
programming languages, including Python, Javascript, C, Golang, and more.

Bindings will be written for Python, our main development language for other
components, and may be written for other languages in the future if deemed
useful.

## Features Planned
- Create data structures, including lists and request queues
- Share these resources over HTTP
- Check requests queues with either non-blocking or blocking calls
- Two APIs: low-level (from C bindings), and high-level (from Python bindings)
- High-level functions in Python bindings for HVAC-specific use case

## Stretch Goals
- Share resources over simpler TCP protocol (eliminate overhead of HTTP)
- Implement security roles in case a component is compromised

## API Documentation

## Low-Level API

This section describes how the bindings for a language interact with the C API.

Examples are provided in Python, and some will reference a `run` function which
can be implemented in Python as follows:

```
# Initialization required
ll = cdll.LoadLibrary("elconn.so") # load the .so library, call it "ll"
ll.elconn_init(0)                  # initialize the library
ii = ll.elconn_make_interpreter()  # make an interpreter, call it "ii"

# The run function
def run(ll, ii, inputList):

  # Convert the input, a Python list, to a JSON string
  strList = json.dumps(inputList)

  # Ask the library to parse JSON and report an ItemID -> List
  listID = ll.elconn_list_from_json(strList.encode())

  # Ask the library to call the interpreter "ii" with the list that was
  # just pased as parameters. (a function name followed by arguments)
  resultID = ll.elconn_call(ii, listID)
```

### Definitions

#### Operation
An operation is a function which can take a list of inputs, and returns a list
of outputs. The reason this is called an "Operation" rather than a "function"
is to make it easier to distinguish between operations which do function-y
things and those which do other things like implement data structures.

- Operations
  - Operations that do function-y things
    - `format`: a function similar to sprintf in C
    - `cat`: a function that concatenates its inputs into a string
  - Operations that do language-y things
    - `: <name> <value>`: assigns `<value>` to `<name>`
    - `store "hello"`: returns an operation which always returns "hello"
    - Any interpreter is also an operation

#### Interpreter

**Practical Definition**

An interpreter is an environment containing keywords. Each keyword either
performs some function, or passes its arguments to a new interpreter with a
different set* of keywords.

*Well, map of keywords, actually

The default interpreter contains the following keywords when created:
- The `:` keyword creates new keywords.
  - Usage: `: <name> <value>`
  - `<value>` should be an Operation type.
    If it is not, the behaviour is undefined.
- The `@` keyword creates a data structure. Data structures are always an
  `Operation` type, so they can be assigned to names using `:`
  - Usage: `@ <type>`
  - Example (JSON): `[":", "mydir", ["@", "directory"]]`
  - Each value for `<type>` is defined under Data Structures
- Builtin functions like `format`, `cat`, and more
  (out of the scope of this document)

**Theory Definition**

An Interpreter is an Operation which takes a keyword as the first parameter,
and passes the remaining parameters to that operation which is named by the
keyword. This is similar to how a LISP interpreter works.


### Data Types

#### ItemID
ItemID is the return value of most functions in the library. This is an unsigned
64-bit integer representing some object (by reference) in the shared library's
memory.

A value of 0 is a null reference. This indicates that an error occurred and no
result could be provided.

#### ItemID -> List
An ItemID can refer to a List. A list contains a number of values of any type.

#### ItemID -> Operation
An ItemID can refer to an Operation. An Operation is any object that can take
a list of parameters as input, perform some function, and return a list.

### Functions

#### `ItemID elconn_init(int32 mode)`

Initialize the library.

Calling this function is required before using any other library function.

The value of `mode` determins the library's behaviour as follows:
- value of `0`: perform normal operation
- value of `1`: perform normal operation, and display debug messages

The return value is the ID of a debug message. (type Debug)

#### `ItemID elconn_list_from_json(string jsonText)`

Parse a JSON string and report an ItemID referring to an object of type List.

This function, and the two following, form the primary API to convert between
data types from this library and data types native to the calling language.

If a JSON object or scalar is passed instead of a list, the result is undefined.
If the JSON could not be parsed, or any other error occurs, 0 is reported.

#### `string elconn_list_to_json(ItemID list)`

(todo: to be implemented soon)

Report a JSON string representing the list referenced by ItemID.

If an error occurs, such as an invalid reference to a list, an string which is
not valid JSON will be returned. A naive way to verfy this, which will work on
the current implementation, is to check if the first character is `[`.

#### `string elconn_list_strfirst(ItemID list)`

(todo: to be implemented soon)

Report the first item in a list as a string.

This is useful if a result is known to be a list containing a JSON string as
its only item, as it avoids the performance loss and general sillyness of
nested JSON encoding.

#### `int32 elconn_list_print(ItemID listToPrint)`

Print a list to STDERR.

0 is returned on success, and -1 is reported if the
ItemID specified does not refer to a list or if any other error occurs.

#### `ItemID elconn_make_interpreter()`

Create an instance of the default interpreter.

#### `ItemID elconn_call(ItemID interpreter, ItemID arguments)`

Execute the operation referenced by `interpreter` using the list of arguments
referenced by `arguments`.

`arguments` typically contains a keyword (a function name) as the first item,
and a list of arguments to that function as the remaining items.

Example (Python):
```
# This will return "Hello, World!" using the function "format"
ll.elconn_call(ii, ["format", "Hello, %s!", "World"])

# Actually, it returns a list with one item: the string "Hello, World!"
```

#### `elconn_serve_remote(string address, ItemID interpreter)`

(todo: specify an error case)

Make an interpreter available over TCP/IP.

#### `ItemID elconn_connect_remote(string address)`

(todo: specify an error case)

Recieve an interpreter that is provided at the specified address.

The ItemID returned refers to a proxy interpreter, which will behave the same
as the provided interpreter does in the host's environment, except that any
call can fail due to a connection error.

### Data Structures

#### `directory`

A directory is an empty interpreter where `:` is used to assign new keywords.

Example (Python):
```
# Create a directory, bind to myCoolDir
run(ll,ii, [":", "myCoolDir", ["@", "directory]] )

# Store the string "hello", bind to "myCoolDir, myCoolStore"
run(ll,ii, ["myCoolDir", ":", "myCoolStore", ["store", "hello"]] )

# Call "myCoolDir, myCoolStore"
valueID = run(ll,ii, ["myCoolDir", "myCoolStore"] )
ll.elconn_list_print(valueID)
```

#### 'requests`

This is an interpreter which implements request queues. The interpreter has
the following keywords:

- `enque`: enqueue an item onto the request queue
  - Usage: `myRequest enque <value>`
- `block`: perform a blocking wait to dequeue an item from the request queue
  - Usage: `myRequest block`
  - Returns a list containing one item
- `check`: Not yet implemented
- `flush`: Not yet implemented

Example (Python with threads):
```
# Create a request queue called myQueue
run(ll,ii, [":", "myQueue", ["@", "requests"]] )

# Schedule something to be enqueued later using a thread
def do_the_thing(ll, ii, item, delay):
    time.sleep(delay)
    run(ll, ii, ["test-map", "b", "enque", item])
thread.start_new_thread(do_the_thing, (ll, ii, "test-value", 4))

print("Wait 4 seconds...")
resID = run(ll, ii, ["test-map", "b", "block"])
ll.elconn_list_print(resID)

```