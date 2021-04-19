package identificator

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrNoIdentifier = errors.New("no identifier")

var identifierContextKey = &contextKey{"identifier context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

func Identificator(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.Split(request.RemoteAddr, ":")
		if len(parts) == 2 {
			// Создаём новый контекст со значением
			ctx := context.WithValue(request.Context(), identifierContextKey, &parts[0])
			// Создаём новый запрос с «заменённым» контекстом:
			request = request.WithContext(ctx)
		}
		// Используем этот запрос далее (в handler'ах).
		handler.ServeHTTP(writer, request)
	})
}

func Identifier(ctx context.Context) (*string, error) {
	value, ok := ctx.Value(identifierContextKey).(*string)
	if !ok {
		return nil, ErrNoIdentifier
	}
	return value, nil
}
