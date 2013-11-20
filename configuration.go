package auwfg

import (
  "time"
  "strings"
)

type Configuration struct {
  address string
  loadRawBody bool
  loadRawQuery bool
  loadRawPath bool
  statsFile string
  bodyPoolSize int
  maxBodySize int64
  invalidPoolSize int
  maxInvalidSize int
  notFound Response
  bodyTooLarge Response
  invalidFormat Response
  statsSleep time.Duration
  internalServerError Response
  routes versions
  dispatcher Dispatcher
  contextFactory ContextFactory
}

type versions map[string]controllers
type controllers map[string]actions
type actions map[string]*Route

func Configure() *Configuration{
  return &Configuration{
    routes: make(versions),
    address: "127.0.0.1:4577",
    bodyPoolSize: 1024,
    maxBodySize: 32769,
    loadRawBody: false,
    loadRawQuery: false,
    invalidPoolSize: 1024,
    maxInvalidSize: 32769,
    statsFile: "stats.json",
    statsSleep: time.Minute,
    dispatcher: genericDispatcher,
    contextFactory: genericContextFactory,
    notFound: NotFound,
    internalServerError: InternalServerError,
    bodyTooLarge: Json(`{"error":"body too large","code":413}`).Status(413).Response,
    invalidFormat: Json(`{"error":"invalid input format","code":400}`).Status(400).Response,
  }
}

func (c *Configuration) Address(address string) *Configuration {
  c.address = address
  return c
}

func (c *Configuration) NotFoundResponse(r Response) *Configuration {
  c.notFound = r
  return c
}

func (c *Configuration) NotFound(body string) *Configuration {
  return c.NotFoundResponse(Json(body).Status(404).Response)
}

func (c *Configuration) BodyTooLargeResponse(r Response) *Configuration {
  c.bodyTooLarge = r
  return c
}

func (c *Configuration) BodyTooLarge(body string) *Configuration {
  return c.BodyTooLargeResponse(Json(body).Status(413).Response)
}

func (c *Configuration) InvalidFormatResponse(r Response) *Configuration {
  c.invalidFormat = r
  return c
}

func (c *Configuration) StatsSleep(sleep time.Duration) *Configuration {
  c.statsSleep = sleep
  return c
}

func (c *Configuration) StatsFile(file string) *Configuration {
  c.statsFile = file
  return c
}

func (c *Configuration) InvalidFormat(body string) *Configuration {
  return c.InvalidFormatResponse(Json(body).Status(400).Response)
}

func (c *Configuration) InternalServerErrorResponse(r Response) *Configuration {
  c.internalServerError = r
  return c
}

func (c *Configuration) InternalServerError(body string) *Configuration {
  return c.InternalServerErrorResponse(Json(body).Status(500).Response)
}

func (c *Configuration) Dispatcher(d Dispatcher) *Configuration {
  c.dispatcher = d
  return c
}

func (c *Configuration) ContextFactory(cf ContextFactory) *Configuration {
  c.contextFactory = cf
  return c
}

func (c *Configuration) BodyPool(poolSize int, bufferSize int) *Configuration {
  c.bodyPoolSize = poolSize
  c.maxBodySize = int64(bufferSize + 1)
  return c
}

func (c *Configuration) InvalidPool(poolSize int, bufferSize int) *Configuration {
  c.invalidPoolSize = poolSize
  c.maxInvalidSize = bufferSize
  return c
}

func (c *Configuration) LoadRawBody() *Configuration {
  c.loadRawBody = true
  return c
}

func (c *Configuration) LoadRawQuery() *Configuration {
  c.loadRawQuery = true
  return c
}

func (c *Configuration) LoadRawPath() *Configuration {
  c.loadRawPath = true
  return c
}

func (c *Configuration) Route(r *RouteBuilder) *Configuration {
  r.method = strings.ToUpper(r.method)
  r.version = strings.ToLower(r.version)
  fullResource := strings.ToLower(r.resource)
  if len(r.parentResource) > 0 {
    fullResource = strings.ToLower(r.parentResource) + ":" + fullResource
  }
  if _, exists := c.routes[r.version]; !exists { c.routes[r.version] = make(controllers) }
  if _, exists := c.routes[r.version][fullResource]; !exists { c.routes[r.version][fullResource] = make(actions) }
  c.routes[r.version][fullResource][r.method] = &Route{
    Action: r.action,
    BodyFactory: r.bf,
  }
  return c
}
