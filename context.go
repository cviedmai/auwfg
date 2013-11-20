package auwfg

import (
  "net"
  "net/http"
)

type ContextFactory func(*BaseContext) interface{}
func genericContextFactory(c *BaseContext) interface{} { return c }

type BaseContext struct {
  Route *Route
  Params *Params
  RawBody []byte
  RawQuery string
  RawPath string
  Body interface{}
  Req *http.Request
  Query map[string]string
  RemoteIp net.IP
}

func newBaseContext(route *Route, params *Params, req *http.Request) *BaseContext{
  return &BaseContext{Route: route, Params: params, Req: req,}
}
