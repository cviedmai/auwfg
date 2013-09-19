package auwfg

import(
  "log"
  "strings"
  "net/url"
  "net/http"
  "encoding/json"
  "github.com/viki-org/bytepool"
)

type Router struct {
  *Configuration
  bodyPool *bytepool.Pool
}

func newRouter(c *Configuration) *Router {
  bp := bytepool.New(c.bodyPoolSize, int(c.maxBodySize))
  return &Router{c, bp}
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
  route, params := r.loadRouteAndParams(req)
  if route == nil {
    r.reply(writer, r.notFound, req)
    return
  }
  bc := newBaseContext(route, params, req)
  if body, res := r.loadBody(route, req); res != nil {
    r.reply(writer, res, req)
    return
  } else {
    bc.Body = body
  }
  bc.Query = loadQuery(req.URL.RawQuery)
  context := r.contextFactory(bc)
  r.reply(writer, r.dispatcher(route, context), req)
}

func (r *Router) reply(writer http.ResponseWriter, res Response, req *http.Request) {
  if res == nil {
    log.Printf("%q \t nil response", req.URL.String())
    res = r.internalServerError
  } else  if res.Status() == 500 {
    if fatal, ok := res.(*FatalResponse); ok {
      log.Printf("%q \t %v", req.URL.String(), fatal.err)
    } else {
      log.Printf("%q \t 500", req.URL.String())
    }
  }

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
  if n, _ := buffer.ReadFrom(req.Body); n == 0 {
    return body, nil
  } else if n == r.maxBodySize {
    return nil, r.bodyTooLarge
  }
  if err := json.Unmarshal(buffer.Bytes(), body); err != nil {
    println(err.Error())
    return nil, r.invalidFormat
  }
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


func loadQuery(raw string) map[string]string {
  l := len(raw)
  if l == 0 { return nil }

  query := make(map[string]string)
  for i := 0; i < l; i++ {
    for ;raw[i] == '&' ;i++ { }
    start := i
    for ;;i++ {
      if i == l { return query }
      if raw[i] == '=' { break; }
    }
    key := raw[start:i]
    i++
    start = i
    for ;;i++ {
      if i == l || raw[i] == '&' ||  raw[i] == '?' { break; }
    }
    value := raw[start:i]
    if escaped, err := url.QueryUnescape(value); err == nil {
      query[strings.ToLower(key)] = escaped
    }
  }
  return query
}
