## TODO List for HA/Connective

### Completed
- [done] elconn_list_from_json
  elconn_list_from_phsh
- [done] elconn_make_interpreter() operation
- [done] elconn_call(operation, list) list
- [done] elconn_serve_remote(port, operation)
- [done] elconn_connect_remote(port) operation

### Next
- elconn_set_exec
- elconn_create() id        // shorthand for elconn_call(":", ...)
                               but also returns operation instead of list
- (: -name -operation)
- (new dir) -operation

- : sa-clusters (new dir)   // makes a function with scope api
- sa-clusters @req sensor0  // makes a request queue called sensor0

### Later

### Misc. Notes

- `:` means add an operation
- `$` means get literal operation
- `&` to take any list/operation and expose it to library users as an id

#### Syntax ideas

##### Creating things
```
: /sa-clusters (new dir)
/sa-clusters : sensor0 (new req-queue)
```

##### Permission masks for functions
```
: public (copy-exec)
within public /sa-clusters (
    : sensor0 (new opmask (list: get) ($ sensor0))
)
```

##### Loops
```
each name (public /sa-clusters list) (
  within public /sa-clusters (
    : (name) (new opmask (list: get) ($ (name)))
  )
)
```

##### Deprecated Notes`
Interface API (@)
- (link-operation name id)
- (: type name) @            // make a sub-api

- : /sa-clusters (new dir)
- @ link-operation name id
- :dir /sa-clusters
- :

