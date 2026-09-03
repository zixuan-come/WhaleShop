package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type ctxKey string

const projectIDKey ctxKey = "pid"

// Project 从 X-Project-Id header 提取项目号,注入 context;缺失/非法都回落 1。
// 呼应 WhaleTestPro 的多项目隔离设计。
func Project(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pid := 1
		if raw := r.Header.Get("X-Project-Id"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 {
				pid = n
			}
		}
		ctx := context.WithValue(r.Context(), projectIDKey, pid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ProjectID(ctx context.Context) int {
	if v, ok := ctx.Value(projectIDKey).(int); ok {
		return v
	}
	return 1
}
