package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/tuxer/go-db"

	"gitlab.com/tuxer/go-json"
)

//JWTAuth ...
type JWTAuth struct {
	Interface
	Query string
}

//Type ...
func (j *JWTAuth) Type() string {
	return TypeJWT
}

//Authenticate ...
func (j *JWTAuth) Authenticate(tx *db.Tx, r *http.Request) error {
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
	rs, e := tx.Get(j.Query, jsPayload.GetString(`iss`))
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
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(jwt[0] + `.` + jwt[1]))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if sig != jwt[2] {
		return fmt.Errorf(`different signature %s %s`, sig, jwt[2])
	}
	return nil
}
