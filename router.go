package auwfg

import(
  "strings"
  "net/http"
  "encoding/json"
  "github.com/viki-org/bytepool"
)

type Router struct {
  *Configuration
  bodyPool *bytepool.Pool
}

func newRouter(c *Configuration) *Router {
  bp := bytepool.New(c.bodyPoolSize, c.maxBodySize)
  return &Router{c, bp}
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
  route, params := r.loadRouteAndParams(req)
  if route == nil {
    reply(writer, r.notFound)
    return
  }
  bc := newBaseContext(route, params, req)
  if body, res := r.loadBody(route, req); res != nil {
    reply(writer, res)
    return
  } else {
    bc.Body = body
  }
  context := r.contextFactory(bc)
  reply(writer, r.dispatcher(route, context))
}

func reply(writer http.ResponseWriter, res Response) {
  h := writer.Header()
  for k, v := range res.Header() { h[k] = v }
  writer.WriteHeader(res.Status())
  writer.Write(res.Body())
}

func (r *Router) loadRouteAndParams(req *http.Request) (*Route, *Params) {
  path := req.URL.Path
  if len(path) < 4 { return nil, nil }

  end := strings.LastIndex(path, ".")
  if end == -1 { end = len(path) }

  parts := strings.Split(path[1:end], "/")
  l := len(parts)
  if l < 2 || l > 5 { return nil, nil }

  for index, part := range parts {
    parts[index] = strings.ToLower(part)
  }

  version, exists := r.routes[parts[0]]
  if exists == false { return nil, nil }

  params := loadParams(parts[1:])
  controller, exists := version[params.Resource]
  if exists == false { return nil, nil }

  m := req.Method
  if m == "GET" && len(params.Id) == 0 { m = "LIST" }

  route, exists := controller[m]
  if exists == false { return nil, nil }

  params.Version = parts[0]
  return route, params
}
func (r *Router) loadBody(route *Route, req *http.Request) (interface{}, Response) {
  defer req.Body.Close()
  if route.BodyFactory == nil { return nil, nil }

  body := route.BodyFactory()
  buffer := r.bodyPool.Checkout()
  defer buffer.Close()
  if n, _ := buffer.ReadFrom(req.Body); n == 0 { return body, nil }
  if err := json.Unmarshal(buffer.Bytes(), body); err != nil { return nil, r.invalidFormat }
  return body, nil
}

func loadParams(parts []string) *Params {
  params := new(Params)
  switch len(parts) {
  case 1:
    params.Resource = parts[0]
  case 2:
    params.Resource = parts[0]
    params.Id = parts[1]
  case 3:
    params.ParentResource = parts[0]
    params.ParentId = parts[1]
    params.Resource = parts[2]
  case 4:
    params.ParentResource = parts[0]
    params.ParentId = parts[1]
    params.Resource = parts[2]
    params.Id = parts[3]
  }
  return params
}
