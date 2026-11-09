// 本文件处理浏览器跨域。OPTIONS 在这里直接返回 204，不会进到登录接口。

package server

import (
	nethttp "net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Filter 给所有响应补上允许任意来源的跨域头，并直接用 204 应答 OPTIONS。
// 允许的请求头里有 Authorization，否则登录后的询价会被浏览器拦下。
// 中间件只在路由匹配成功后才运行，所以预检不能只写在中间件里。
func Filter() khttp.FilterFunc {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,PATCH,DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,Accept")
			if r.Method == nethttp.MethodOptions {
				w.WriteHeader(nethttp.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
