// Pulled from https://github.com/youtube/vitess 229422035ca0c716ad0c1397ea1351fe62b0d35a
// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package czlib

// See http://www.zlib.net/zlib_how.html for more information on this

/*
#cgo CFLAGS: -O3 -Werror=implicit
#cgo pkg-config: zlib

#include "zlib.h"

// inflateInit is a macro, so using a wrapper function
int zstream_inflate_init(char *strm) {
  ((z_stream*)strm)->zalloc = Z_NULL;
  ((z_stream*)strm)->zfree = Z_NULL;
  ((z_stream*)strm)->opaque = Z_NULL;

  ((z_stream*)strm)->next_in = Z_NULL;
  ((z_stream*)strm)->avail_in = 0;
  ((z_stream*)strm)->total_in = 0;

  ((z_stream*)strm)->next_out = Z_NULL;
  ((z_stream*)strm)->avail_out = 0;
  ((z_stream*)strm)->total_out = 0;

  return inflateInit((z_stream*)strm);
}

// deflateInit is a macro, so using a wrapper function
int zstream_deflate_init(char *strm, int level) {
  ((z_stream*)strm)->zalloc = Z_NULL;
  ((z_stream*)strm)->zfree = Z_NULL;
  ((z_stream*)strm)->opaque = Z_NULL;
  return deflateInit((z_stream*)strm, level);
}

unsigned int zstream_avail_in(char *strm) {
  return ((z_stream*)strm)->avail_in;
}

unsigned int zstream_avail_out(char *strm) {
  return ((z_stream*)strm)->avail_out;
}

char* zstream_msg(char *strm) {
  return ((z_stream*)strm)->msg;
}

void zstream_set_in_buf(char *strm, void *buf, unsigned int len) {
  ((z_stream*)strm)->next_in = (Bytef*)buf;
  ((z_stream*)strm)->avail_in = len;
}

void zstream_set_out_buf(char *strm, void *buf, unsigned int len) {
  ((z_stream*)strm)->next_out = (Bytef*)buf;
  ((z_stream*)strm)->avail_out = len;
}

unsigned int zstream_total_in(char *strm) {
  return ((z_stream*)strm)->total_in;
}

unsigned int zstream_total_out(char *strm) {
  return ((z_stream*)strm)->total_out;
}

void zstream_set_total_in(char *strm, unsigned long total) {
	((z_stream*)strm)->total_in = total;
}

void zstream_set_total_out(char *strm, unsigned long total) {
	((z_stream*)strm)->total_out = total;
}

int zstream_inflate(char *strm, int flag) {
  return inflate((z_stream*)strm, flag);
}

int zstream_deflate(char *strm, int flag) {
  return deflate((z_stream*)strm, flag);
}

void zstream_inflate_end(char *strm) {
  inflateEnd((z_stream*)strm);
}

void zstream_deflate_end(char *strm) {
  deflateEnd((z_stream*)strm);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	zNoFlush   = C.Z_NO_FLUSH
	zSyncFlush = C.Z_SYNC_FLUSH
)

// z_stream is a buffer that's big enough to fit a C.z_stream.
// This lets us allocate a C.z_stream within Go, while keeping the contents
// opaque to the Go GC. Otherwise, the GC would look inside and complain that
// the pointers are invalid, since they point to objects allocated by C code.
type Zstream [unsafe.Sizeof(C.z_stream{})]C.char

func NewZstream() (*Zstream, error) {
	var strm Zstream
	err := strm.inflateInit()
	return &strm, err
}

func (strm *Zstream) inflateInit() error {
	result := C.zstream_inflate_init(&strm[0])
	if result != Z_OK {
		return fmt.Errorf("cgzip: failed to initialize inflate (%v): %v", result, strm.msg())
	}
	fmt.Println("inflate init ok")
	return nil
}

func (strm *Zstream) deflateInit(level int) error {
	result := C.zstream_deflate_init(&strm[0], C.int(level))
	if result != Z_OK {
		return fmt.Errorf("cgzip: failed to initialize deflate (%v): %v", result, strm.msg())
	}
	return nil
}

func (strm *Zstream) inflateEnd() {
	C.zstream_inflate_end(&strm[0])
}

func (strm *Zstream) deflateEnd() {
	C.zstream_deflate_end(&strm[0])
}

func (strm *Zstream) availIn() int {
	return int(C.zstream_avail_in(&strm[0]))
}

func (strm *Zstream) availOut() int {
	return int(C.zstream_avail_out(&strm[0]))
}

func (strm *Zstream) msg() string {
	return C.GoString(C.zstream_msg(&strm[0]))
}

func (strm *Zstream) setInBuf(buf []byte, size int) {
	if buf == nil {
		C.zstream_set_in_buf(&strm[0], nil, C.uint(size))
	} else {
		C.zstream_set_in_buf(&strm[0], unsafe.Pointer(&buf[0]), C.uint(size))
	}
}

func (strm *Zstream) setOutBuf(buf []byte, size int) {
	if buf == nil {
		C.zstream_set_out_buf(&strm[0], nil, C.uint(size))
	} else {
		C.zstream_set_out_buf(&strm[0], unsafe.Pointer(&buf[0]), C.uint(size))
	}
}

func (strm *Zstream) totalIn() int {
	return int(C.zstream_total_in(&strm[0]))
}

func (strm *Zstream) totalOut() int {
	return int(C.zstream_total_out(&strm[0]))
}

func (strm *Zstream) setTotalIn(total int) {
	C.zstream_set_total_in(&strm[0], C.ulong(total))
}

func (strm *Zstream) setTotalOut(total int) {
	C.zstream_set_total_out(&strm[0], C.ulong(total))
}

func (strm *Zstream) inflate(flag int) (int, error) {
	ret := C.zstream_inflate(&strm[0], C.int(flag))
	switch ret {
	case Z_NEED_DICT:
		ret = Z_DATA_ERROR
		fallthrough
	case Z_DATA_ERROR, Z_MEM_ERROR:
		return int(ret), fmt.Errorf("cgzip: failed to inflate (%v): %v", ret, strm.msg())
	}
	return int(ret), nil
}

func (strm *Zstream) deflate(flag int) {
	ret := C.zstream_deflate(&strm[0], C.int(flag))
	if ret == Z_STREAM_ERROR {
		// all the other error cases are normal,
		// and this should never happen
		panic(fmt.Errorf("cgzip: Unexpected error (1)"))
	}
}

func (strm *Zstream) Reset() {
	strm.setInBuf(nil, 0)
	strm.setOutBuf(nil, 0)
}
