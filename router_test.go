package auwfg

import (
  "testing"
  "net/http/httptest"
  "github.com/viki-org/gspec"
)

func TestNotFoundOnShortUrls(t *testing.T) {
  req := gspec.Request().Url("/x").Req
  res := httptest.NewRecorder()
  newRouter(Configure()).ServeHTTP(res, req)
  assertResponse(t, res, 404, `{"error":"not found","code":404}`)
}

func TestNotFoundOnUnknownVersions(t *testing.T) {
  req := gspec.Request().Url("/v3/sessions.json").Req
  res := httptest.NewRecorder()
  newRouter(Configure()).ServeHTTP(res, req)
  assertResponse(t, res, 404, `{"error":"not found","code":404}`)
}

func TestNotFoundOnUnknownControllers(t *testing.T) {
  req := gspec.Request().Url("/v4/cats.json").Req
  res := httptest.NewRecorder()
  newRouter(Configure()).ServeHTTP(res, req)
  assertResponse(t, res, 404, `{"error":"not found","code":404}`)
}

func TestNotFoundOnUnknownActions(t *testing.T) {
  req := gspec.Request().Url("/v4/sessions.json").Req
  res := httptest.NewRecorder()
  newRouter(Configure()).ServeHTTP(res, req)
  assertResponse(t, res, 404, `{"error":"not found","code":404}`)
}

func TestExecutesAListAction(t *testing.T) {
  f := func(context interface{}) Response {return Json(`{"spice":"mustflow"}`).Response}
  req := gspec.Request().Url("/v4/sessions.json").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("LIST", "v4", "sessions", f)))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"spice":"mustflow"}`)
}

func TestExecutesAGetAction(t *testing.T) {
  f := func(context interface{}) Response { return Json(`{"name":"duncan"}`).Response }
  req := gspec.Request().Url("/v2/gholas/123g.json").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("GET", "v2", "gholas", f)))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"name":"duncan"}`)
}

func TestExecutesAnActionRegardlessOfCasing(t *testing.T) {
  f := func(context interface{}) Response { return Json(`{"name":"duncan"}`).Response }
  req := gspec.Request().Url("/v2/GHOLAS/123g.json").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("GET", "v2", "gholas", f)))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"name":"duncan"}`)
}

func TestExecutesAPostAction(t *testing.T) {
  f := func(context interface{}) Response { return Json(`{"name":"shaihulud"}`).Response }
  req := gspec.Request().Url("/v1/worms.json").Method("POST").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("POST", "v1", "worms", f)))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"name":"shaihulud"}`)
}

func TestExecutesAPutAction(t *testing.T) {
  f := func(context interface{}) Response { return Json(`{"name":"shaihulud"}`).Response }
  req := gspec.Request().Url("/v1/worms/22w.json").Method("PUT").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("PUT", "v1", "worms", f)))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"name":"shaihulud"}`)
}

func TestRoutesToANestedResource(t *testing.T) {
  spec := gspec.New(t)
  f := func(context interface{}) Response {
    spec.Expect(context.(*BaseContext).Params.ParentResource).ToEqual("gholas")
    spec.Expect(context.(*BaseContext).Params.ParentId).ToEqual("123g")
    spec.Expect(context.(*BaseContext).Params.Resource).ToEqual("history")
    spec.Expect(context.(*BaseContext).Params.Id).ToEqual("")
    return Json(`{"name":"history"}`).Response
  }
  req := gspec.Request().Url("/v2/gholas/123g/history.json").Req
  res := httptest.NewRecorder()
  router := newRouter(Configure().Route(R("LIST", "v2", "history", f).Parent("gholas")))
  router.ServeHTTP(res, req)
  assertResponse(t, res, 200, `{"name":"history"}`)
}

func TestCreatingAndDispatchingThroughCustomTypes(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Name).ToEqual("leto")
    spec.Expect(context.Params.Version).ToEqual("v1")
    spec.Expect(context.Params.Resource).ToEqual("worms")
    spec.Expect(context.Params.Id).ToEqual("22w")
    return Json(`{"name":"bigjohn"}`).Status(203).Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").Req
  res := httptest.NewRecorder()
  newRouter(c).ServeHTTP(res, req)
  assertResponse(t, res, 203, `{"name":"bigjohn"}`)
}

func TestDefaulsTheBodyValueWhenNoBodyIsPresent(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    input := context.Body.(*TestBody)
    spec.Expect(input.Hello).ToEqual("")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f).BodyFactory(testBodyFactory)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestParsesABody(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    input := context.Body.(*TestBody)
    spec.Expect(input.Hello).ToEqual("World")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f).BodyFactory(testBodyFactory)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").BodyString(`{"hello":"World"}`).Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestDoesNotStoreRawBodyWhenNotConfigured(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(string(context.RawBody)).ToEqual("")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").BodyString(`{"hello":"World"}`).Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestStoresRawBody(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(string(context.RawBody)).ToEqual(`{"hello":"World"}`)
    return Json("").Response
  }
  c := Configure().LoadRawBody().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").BodyString(`{"hello":"World"}`).Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestHandlesANilResponse(t *testing.T) {
  f := func(context interface{}) Response { return nil }
  c := Configure().Route(R("GET", "v1", "worms", f))
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").Req
  res := httptest.NewRecorder()
  newRouter(c).ServeHTTP(res, req)
  assertResponse(t, res, 500, `{"error":"internal server error","code":500}`)
}

func TestHandlesBodiesLargerThanAllowed(t *testing.T) {
  f := func(context *TestContext) Response { return Json("").Response }
  c := Configure().Route(R("GET", "v1", "worms", f).BodyFactory(testBodyFactory)).ContextFactory(testContextFactory).Dispatcher(testDispatcher).BodyPool(3, 1)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").BodyString(`{"hello":"World"}`).Req
  res := httptest.NewRecorder()
  newRouter(c).ServeHTTP(res, req)
  assertResponse(t, res, 413, `{"error":"body too large","code":413}`)
}

func TestParsesQueryString(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Query["app"]).ToEqual("6003")
    spec.Expect(context.Query["t"]).ToEqual("1 2")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?APP=6003&t=1%202&").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestDoesNotStoreRawQueryWhenNotConfigured(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.RawQuery).ToEqual("")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").BodyString(`{"hello":"World"}`).Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestStoresRawQuery(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.RawQuery).ToEqual("app=121&gogo=yes")
    return Json("").Response
  }
  c := Configure().LoadRawQuery().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?app=121&gogo=yes").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestParsesQueryStirngWithEmptyPairAtTheStart(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Query["app"]).ToEqual("100004a")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?&app=100004a").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestParsesQueryStirngWithEmptyPairInTheMiddle(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Query["app"]).ToEqual("100004a")
    spec.Expect(context.Query["t"]).ToEqual("1")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?app=100004a&&t=1").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestHandlesMultipleQuestionMarksInQueryString(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Query["app"]).ToEqual("100005a")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?app=100002a?app=100005a").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestParsesAQueryStringWithAMissingValue(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(len(context.Query)).ToEqual(0)
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?a").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestParsesAQueryStringWithAMissingValue2(t *testing.T) {
  spec := gspec.New(t)
  f := func(context *TestContext) Response {
    spec.Expect(context.Query["b"]).ToEqual("")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json?b=").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestNoIpPresent(t *testing.T) {
  f := func(context *TestContext) Response {
    gspec.New(t).Expect(context.RemoteIp).ToBeNil()
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestInvalidIp(t *testing.T) {
  f := func(context *TestContext) Response {
    gspec.New(t).Expect(context.RemoteIp).ToBeNil()
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Url("/v1/worms/22w.json").Header("x-forwarded-for", "12.12").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestLoadIpFromXForwardedIp(t *testing.T) {
  f := func(context *TestContext) Response {
    gspec.New(t).Expect(context.RemoteIp.String()).ToEqual("12.12.12.12")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Header("x-forwarded-for", "12.12.12.12,13.13.13.13").Header("client-ip", "11.11.11.11")
  req = req.Header("remote-addr", "10.10.10.10").Url("/v1/worms/22w.json").Method("GET")
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req.Req)
}

func TestLoadIpFromClientIp(t *testing.T) {
  f := func(context *TestContext) Response {
    gspec.New(t).Expect(context.RemoteIp.String()).ToEqual("10.10.10.10")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Header("client-ip", "10.10.10.10").Header("remote-addr", "12.12.10.10").Url("/v1/worms/22w.json").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func TestLoadIpFromRemoteAddr(t *testing.T) {
  f := func(context *TestContext) Response {
    gspec.New(t).Expect(context.RemoteIp.String()).ToEqual("13.10.10.10")
    return Json("").Response
  }
  c := Configure().Route(R("GET", "v1", "worms", f)).ContextFactory(testContextFactory).Dispatcher(testDispatcher)
  req := gspec.Request().Header("remote-addr", "13.10.10.10").Url("/v1/worms/22w.json").Method("GET").Req
  newRouter(c).ServeHTTP(httptest.NewRecorder(), req)
}

func assertResponse(t *testing.T, res *httptest.ResponseRecorder, status int, raw string) {
  spec := gspec.New(t)
  spec.Expect(res.Code).ToEqual(status)
  spec.Expect(string(res.Body.Bytes())).ToEqual(raw)
}

type TestContext struct {
  Name string
  *BaseContext
}

type TestBody struct {
  Hello string
}

func testContextFactory(c *BaseContext) interface{} {
  return &TestContext{"leto", c}
}

func testDispatcher(route *Route, context interface{}) Response {
  return route.Action.(func(*TestContext) Response)(context.(*TestContext))
}

func testBodyFactory() interface{} {
  return new(TestBody)
}
