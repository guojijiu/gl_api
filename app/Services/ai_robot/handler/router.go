package handler

import "fmt"

type RouteKey struct {
	Platform     string
	QuestionType int
}

type HandlerFunc func() error

// Router 负责意图处理前的平台+业务类型分发。
// 只做路由职责，具体业务逻辑由注册的 HandlerFunc 完成。
type Router struct {
	routes map[RouteKey]HandlerFunc
}

func NewRouter() *Router {
	return &Router{routes: map[RouteKey]HandlerFunc{}}
}

func (r *Router) Register(key RouteKey, fn HandlerFunc) {
	if fn == nil {
		return
	}
	r.routes[key] = fn
}

func (r *Router) Dispatch(key RouteKey) error {
	fn := r.routes[key]
	if fn == nil {
		return fmt.Errorf("no handler for platform=%s question_type=%d", key.Platform, key.QuestionType)
	}
	return fn()
}
