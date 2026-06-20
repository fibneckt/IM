package handler

import (
	"IM/apps/im/ws/internal/svc"
	"IM/pkg/ctxdata"
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/token"
)

type JwtAuth struct {
	svc    *svc.ServiceContext
	parser *token.TokenParser
	logx.Logger
}

func (j JwtAuth) Auth(w http.ResponseWriter, r *http.Request) bool {
<<<<<<< HEAD
=======
	// j.Infof("request headers: %v", r.Header) // 加这行看看到底有没有 Authorization
>>>>>>> 72ed3df (修正了 JwtSecret 不一致导致的错误)
	tok, err := j.parser.ParseToken(r, j.svc.Config.JwtAuth.AccessSecret, "")
	if err != nil {
		// 解析错误
		j.Errorf("parse token err: %v", err)
		return false
	}

	if !tok.Valid {
		return false
	}

	// 完成 jwt 验证
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// 把 id 放到 request 中
	*r = *r.WithContext(context.WithValue(r.Context(), ctxdata.Identify, claims[ctxdata.Identify]))
	return true
}

func (j JwtAuth) UserId(r *http.Request) string {
	return ctxdata.GetUid(r.Context())
}

func NewJwtAuth(svc *svc.ServiceContext) *JwtAuth {
	return &JwtAuth{
		svc:    svc,
		parser: token.NewTokenParser(),
		Logger: logx.WithContext(context.Background()),
	}
}
