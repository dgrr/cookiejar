package cookiejar

import (
	"io"
	"sync"
	"unsafe"

	"github.com/valyala/fasthttp"
)

var cookiePool = sync.Pool{
	New: func() interface{} {
		return &CookieJar{}
	},
}

// AcquireCookieJar returns an empty CookieJar object from pool
func AcquireCookieJar() *CookieJar {
	return cookiePool.Get().(*CookieJar)
}

// ReleaseCookieJar returns CookieJar to the pool
func ReleaseCookieJar(c *CookieJar) {
	c.Release()
	cookiePool.Put(c)
}

// CookieJar is container of cookies
//
// This object is used to handle multiple cookies
type CookieJar map[string]*fasthttp.Cookie

// Set sets cookie using key-value
//
// This function can replace an existent cookie
func (cj *CookieJar) Set(key, value string) {
	setCookie(cj, key, value)
}

// Get returns and delete a value from cookiejar.
func (cj *CookieJar) Get() *fasthttp.Cookie {
	for k, v := range *cj {
		delete(*cj, k)
		return v
	}
	return nil
}

// SetBytesK sets cookie using key=value
//
// This function can replace an existent cookie.
func (cj *CookieJar) SetBytesK(key []byte, value string) {
	setCookie(cj, b2s(key), value)
}

// SetBytesV sets cookie using key=value
//
// This function can replace an existent cookie.
func (cj *CookieJar) SetBytesV(key string, value []byte) {
	setCookie(cj, key, b2s(value))
}

// SetBytesKV sets cookie using key=value
//
// This function can replace an existent cookie.
func (cj *CookieJar) SetBytesKV(key, value []byte) {
	setCookie(cj, b2s(key), b2s(value))
}

func setCookie(cj *CookieJar, key, value string) {
	c, ok := (*cj)[key]
	if !ok {
		c = fasthttp.AcquireCookie()
	}
	c.SetKey(key)
	c.SetValue(value)
	(*cj)[key] = c
}

// SetCookie sets cookie using its key.
//
// After that you can use Peek or Get function to get cookie value.
func (cj *CookieJar) Put(cookie *fasthttp.Cookie) {
	c, ok := (*cj)[b2s(cookie.Key())]
	if ok {
		fasthttp.ReleaseCookie(c)
	}
	(*cj)[b2s(cookie.Key())] = cookie
}

// Peek peeks cookie value using key.
//
// This function does not delete cookie
func (cj *CookieJar) Peek(key string) *fasthttp.Cookie {
	return (*cj)[key]
}

// Release releases all cookie values.
func (cj *CookieJar) Release() {
	for k := range *cj {
		cj.ReleaseCookie(k)
	}
}

// ReleaseCookie releases a cookie specified by parsed key.
func (cj *CookieJar) ReleaseCookie(key string) {
	c, ok := (*cj)[key]
	if ok {
		fasthttp.ReleaseCookie(c)
		delete(*cj, key)
	}
}

// PeekValue returns value of specified cookie-key.
func (cj *CookieJar) PeekValue(key string) []byte {
	c, ok := (*cj)[key]
	if ok {
		return c.Value()
	}
	return nil
}

// ResponseCookies gets all Response cookies reading Set-Cookie header.
func (cj *CookieJar) ReadResponse(r *fasthttp.Response) {
	r.Header.VisitAllCookie(func(key, value []byte) {
		cookie := fasthttp.AcquireCookie()
		cookie.ParseBytes(value)
		cj.Put(cookie)
	})
}

// RequestCookies gets all cookies from a Request reading Set-Cookie header.
func (cj *CookieJar) RequestRequest(r *fasthttp.Request) {
	r.Header.VisitAllCookie(func(key, value []byte) {
		cookie := fasthttp.AcquireCookie()
		cookie.ParseBytes(value)
		cj.Put(cookie)
	})
}

// WriteTo writes all cookies representation to w.
func (cj *CookieJar) WriteTo(w io.Writer) (n int64, err error) {
	for _, c := range *cj {
		nn, err := c.WriteTo(w)
		n += nn
		if err != nil {
			break
		}
	}
	return
}

// FillRequest dumps all cookies stored in cj into Request adding this values to Cookie header.
func (cj *CookieJar) FillRequest(r *fasthttp.Request) {
	for _, c := range *cj {
		r.Header.SetCookieBytesKV(c.Key(), c.Value())
	}
}

// FillResponse dumps all cookies stored in cj into Response adding this values to Cookie header.
func (cj *CookieJar) FillResponse(r *fasthttp.Response) {
	for _, c := range *cj {
		r.Header.SetCookie(c)
	}
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
