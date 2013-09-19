package auwfg

import (
  "strconv"
  "net/http"
)

var (
  JsonHeader = []string{"application/json; charset=utf-8"}
)

type Response interface {
  Status() int
  Body() []byte
  Header() http.Header
}

type NormalResponse struct {
  S int
  B []byte
  H http.Header
}

func (r *NormalResponse) Status() int {
  return r.S
}

func (r *NormalResponse) Body() []byte {
  return r.B
}

func (r *NormalResponse) Header() http.Header {
  return r.H
}

func JsonResponse(raw string, s int) Response {
  b := []byte(raw)
  return &NormalResponse{
    S: s,
    B: b,
    H: http.Header{"Content-Type": JsonHeader, "Content-Length": []string{strconv.Itoa(len(b))}},
  }
}