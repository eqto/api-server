package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/tuxer/go-json"
)

func jwtGenerate(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func jwtAuthorize(s *Server, r *http.Request) error {
	authorization := r.Header.Get(`Authorization`)
	if len(authorization) < 7 {
		return fmt.Errorf(`invalid authorization token %s`, authorization)
	}
	jwt := strings.Split(authorization[7:], `.`)

	b, e := base64.RawURLEncoding.DecodeString(jwt[1])
	if e != nil {
		return e
	}
	jsPayload := json.Parse(b)
	rs, e := s.cn.Get(s.authQuery, jsPayload.GetString(`iss`))
	if e != nil {
		return e
	}
	if rs == nil {
		return fmt.Errorf(`user %s not found`, jsPayload.GetString(`iss`))
	}
	secret := ``
	for key := range rs {
		secret = rs.String(key)
	}
	sig := jwtGenerate(jwt[0]+`.`+jwt[1], secret)
	if sig != jwt[2] {
		return fmt.Errorf(`different signature %s %s`, sig, jwt[2])
	}
	return nil
}
