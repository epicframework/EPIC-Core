/* Created by "go tool cgo" - DO NOT EDIT. */

/* package command-line-arguments */

/* Start of preamble from import "C" comments.  */









/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt32 GoInt;
typedef GoUint32 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_32_bit_pointer_matching_GoInt[sizeof(void*)==32/8 ? 1:-1];

typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


extern GoInt32 elconn_get_type(GoInt32 p0);

extern void elconn_display_info(GoInt32 p0);

extern GoInt32 elconn_make_interpreter();

extern GoInt32 elconn_call(GoInt32 p0, GoInt32 p1);

extern void elconn_link(char* p0, GoInt32 p1, GoInt32 p2);

extern GoInt32 elconn_list_from_json(char* p0);

extern GoInt32 elconn_list_from_text(char* p0);

extern char* elconn_list_to_json(GoInt32 p0);

extern char* elconn_list_strfirst(GoInt32 p0);

extern GoInt32 elconn_list_print(GoInt32 p0);

extern GoInt32 elconn_init(GoInt32 p0);

extern GoInt32 elconn_serve_remote(char* p0, GoInt32 p1);

extern GoInt32 elconn_connect_remote(char* p0);

#ifdef __cplusplus
}
#endif
