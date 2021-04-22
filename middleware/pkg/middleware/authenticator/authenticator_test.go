package authenticator

import (
	"bytes"
	"context"
	"github.com/go-chi/chi"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticatorHTTPMux(t *testing.T) {
	mux := http.NewServeMux()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USERAUTH", nil
	})

	authenticatorFail := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "FAIL", nil
	})

	mux.Handle(
		"/get",
		authenticatorMd(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				if err == ErrNoAuthentication {
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Fatal(err)
			}
			data := profile.(string)

			if data != "USERAUTH" && request.RemoteAddr != "192.0.2.1" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	mux.Handle(
		"/post",
		authenticatorFail(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				if err == ErrNoAuthentication {
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Fatal(err)
			}
			data := profile.(string)

			if data != "USERAUTH" && request.RemoteAddr != "192.0.2.1" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		})),
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name     string
		args     args
		want     []byte
		wantCode int
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "192.0.2.1"}, want: []byte("USERAUTH"), wantCode: http.StatusOK},
		// TODO: write for other methods
		{name: "No auth", args: args{method: "POST", path: "/post", addr: "191.0.2.1"}, want: []byte{}, wantCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("got %s, want %s", got, tt.want)
		}
		gotCode := response.Code
		if gotCode != tt.wantCode {
			t.Errorf("got %d, want %d", gotCode, tt.wantCode)
		}
	}
}

func TestAuthenticatorChi(t *testing.T) {
	router := chi.NewRouter()
	authenticatorMd := Authenticator(func(ctx context.Context) (*string, error) {
		id := "192.0.2.1"
		return &id, nil
	}, func(ctx context.Context, id *string) (interface{}, error) {
		return "USERAUTH", nil
	})
	router.With(authenticatorMd).Get(
		"/get",
		func(writer http.ResponseWriter, request *http.Request) {
			profile, err := Authentication(request.Context())
			if err != nil {
				if err == ErrNoAuthentication {
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Fatal(err)
			}
			data := profile.(string)

			if data == "USERAUTH" && request.RemoteAddr != "192.0.2.1" {
				writer.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, err = writer.Write([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
		},
	)

	type args struct {
		method string
		path   string
		addr   string
	}

	tests := []struct {
		name     string
		args     args
		want     []byte
		wantCode int
	}{
		{name: "GET", args: args{method: "GET", path: "/get", addr: "191.0.2.1"}, want: []byte("USERAUTH"), wantCode: http.StatusOK},
		// TODO: write for other methods
		{name: "No auth", args: args{method: "GET", path: "/get", addr: "191.0.0.1"}, want: []byte{}, wantCode: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
		request.RemoteAddr = tt.args.addr
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		got := response.Body.Bytes()
		if !bytes.Equal(tt.want, got) {
			t.Errorf("%s, got %s, want %s", tt.name, got, tt.want)
		}
		gotCode := response.Code
		if tt.wantCode != gotCode {
			t.Errorf("%s, got %d, want %d", tt.name, gotCode, tt.wantCode)
		}
	}
}
