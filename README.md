# cookiejar

CookieJar for [fasthttp](https://github.com/valyala/fasthttp).

# Example

```go
package main

import (
	"fmt"

	"github.com/dgrr/cookiejar"
	"github.com/valyala/fasthttp"
)

func main() {
	server := fasthttp.Server{
		Handler: handler,
	}
	go server.ListenAndServe(":8080")

	doRequest("http://localhost:8080")
	server.Shutdown()
}

func handler(ctx *fasthttp.RequestCtx) {
	// Acquire cookie jar
	cj := cookiejar.AcquireCookieJar()
	defer cookiejar.ReleaseCookieJar(cj)

	// filling cookiejar
	cj.Set("Hello", "world")
	cj.Set("make", "fasthttp")
	cj.Set("great", "again")

	// writing values to the response
	cj.FillResponse(&ctx.Response)
}

func doRequest(addr string) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	// Acquire cookie jar
	cj := cookiejar.AcquireCookieJar()
	defer cookiejar.ReleaseCookieJar(cj)

	req.SetRequestURI(addr)
	req.SetConnectionClose()
	err := fasthttp.Do(req, res)
	if err != nil {
		panic(err)
	}
	// Read cookies from the response
	cj.ReadResponse(res)

	for {
		// Read cookie by cookie
		c := cj.Get()
		if c == nil {
			break
		}
		fmt.Printf("Collected cookies: %s=%s\n", c.Key(), c.Value())
		fasthttp.ReleaseCookie(c)
	}
}
```
