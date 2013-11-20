package auwfg

import(
  "log"
  "net"
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
  Stats.Request()
  route, params := r.loadRouteAndParams(req)
  if route == nil {
    r.reply(writer, r.notFound, req)
    return
  }
  bc := newBaseContext(route, params, req)
  if res := r.loadBody(route, req, bc); res != nil {
    r.reply(writer, res, req)
    return
  }

  bc.Query = loadQuery(req.URL.RawQuery)
  if r.loadRawQuery {
    bc.RawQuery = req.URL.RawQuery
  }

  bc.RemoteIp = loadRemoteIp(req)
  context := r.contextFactory(bc)
  r.reply(writer, r.dispatcher(route, context), req)
}

func (r *Router) reply(writer http.ResponseWriter, res Response, req *http.Request) {
  if res == nil {
    log.Printf("%q nil response", req.URL.String())
    res = r.internalServerError
  } else {
    status := res.GetStatus()
    if status >= 500 {
      Stats.Fatal()
      if fatal, ok := res.(*FatalResponse); ok {
        log.Printf("%q %v", req.URL.String(), fatal.err)
      } else {
        log.Printf("%q 500", req.URL.String())
      }
    } else if status >= 400 {
      Stats.Error()
    }
  }

  defer res.Close()
  h := writer.Header()
  for k, v := range res.GetHeader() { h[k] = v }
  writer.WriteHeader(res.GetStatus())
  writer.Write(res.GetBody())
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
    parts[index] = part
  }

  version, exists := r.routes[parts[0]]
  if exists == false { return nil, nil }

  params := loadParams(parts[1:])
  var fullResource = params.Resource
  if len(params.ParentResource) > 0  {
    fullResource = params.ParentResource + ":" + fullResource
  }
  controller, exists := version[fullResource]
  if exists == false { return nil, nil }

  m := req.Method
  if m == "GET" && len(params.Id) == 0 { m = "LIST" }

  route, exists := controller[m]
  if exists == false { return nil, nil }

  params.Version = parts[0]
  return route, params
}
func (r *Router) loadBody(route *Route, req *http.Request, context *BaseContext) Response {
  defer req.Body.Close()
  buffer := r.bodyPool.Checkout()
  defer buffer.Close()
  if route.BodyFactory != nil { context.Body = route.BodyFactory() }

  if n, _ := buffer.ReadFrom(req.Body); n == 0 {
    return nil
  } else if n == r.maxBodySize {
    return r.bodyTooLarge
  }
  if r.loadRawBody {
    context.RawBody = make([]byte, buffer.Len())
    copy(context.RawBody, buffer.Bytes())
  }
  if context.Body != nil {
    if err := json.Unmarshal(buffer.Bytes(), context.Body); err != nil {
      return r.invalidFormat
    }
  }
  return nil
}

func loadParams(parts []string) *Params {
  params := new(Params)
  switch len(parts) {
  case 1:
    params.Resource = strings.ToLower(parts[0])
  case 2:
    params.Resource = strings.ToLower(parts[0])
    params.Id = parts[1]
  case 3:
    params.ParentResource = strings.ToLower(parts[0])
    params.ParentId = parts[1]
    params.Resource = parts[2]
  case 4:
    params.ParentResource = strings.ToLower(parts[0])
    params.ParentId = parts[1]
    params.Resource = strings.ToLower(parts[2])
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

func loadRemoteIp(req *http.Request) net.IP {
  //susceptible to ip spoofing attacks. Could implement a whitelist of trusted proxies.
  ips := req.Header.Get("x-forwarded-for")
  if len(ips) == 0 {
    ips = req.Header.Get("client-ip")
    if len(ips) == 0 {
      ips = req.Header.Get("remote-addr")
      if len(ips) == 0 { return nil }
    }
  }
  index := strings.Index(ips, ",")
  if index != -1 { ips = ips[0:index]}
  return net.ParseIP(ips)
}
