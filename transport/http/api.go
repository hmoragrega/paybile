package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hmoragrega/paybile/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// userIDCtxKey context key that holds the user request.
type userIDCtxKey struct{}

type ApiService struct {
	LoginService      LoginService
	TransactionLister TransactionLister
	TransferCreator   TransferCreator
	WalletGetter      WalletGetter
}

func (svc *ApiService) Handler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(middleware.RequestID)
	r.Use(svc.requestIDResponseHeader())
	r.Use(middleware.Timeout(time.Second * 30))
	r.Use(middleware.Heartbeat("/api/v1/health"))
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, notFoundErr, nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, methodNotAllowedErr, nil)
	})

	auth := r.With(svc.authMiddleware())

	auth.Get("/api/v1/wallet/{walletID}", GetWalletHandler(svc.WalletGetter))
	auth.Get("/api/v1/wallet/{walletID}/transactions", TransactionListHandler(svc.TransactionLister))
	auth.Post("/api/v1/wallet/{walletID}/transfer", CreateTransferHandler(svc.TransferCreator))

	return r
}

var (
	readRequestBodyErr = apiError{
		Status: http.StatusBadRequest,
		Key:    "ReadRequestBodyError",
		error:  errors.New("cannot read request body"),
	}
	unmarshalBodyErr = apiError{
		Status: http.StatusBadRequest,
		Key:    "UnmarshalErr",
		error:  errors.New("cannot unmarshal request body"),
	}
	unauthorizedErr = apiError{
		Status: http.StatusUnauthorized,
		Key:    "UnauthorizedErr",
		error:  errors.New("unauthorized"),
	}
	forbiddenErr = apiError{
		Status: http.StatusForbidden,
		Key:    "ForbiddenErr",
		error:  errors.New("forbidden"),
	}
	notFoundErr = apiError{
		Status: http.StatusNotFound,
		Key:    "NotFoundErr",
		error:  errors.New("not found"),
	}
	methodNotAllowedErr = apiError{
		Status: http.StatusNotFound,
		Key:    "MethodNotAllowedErr",
		error:  errors.New("method not allowed"),
	}
	badRequestErr = apiError{
		Status: http.StatusBadRequest,
		Key:    "BadRequestErr",
		error:  errors.New("bad request"),
	}
	unprocessableEntityErr = apiError{
		Status: http.StatusUnprocessableEntity,
		Key:    "UnprocessableEntityErr",
		error:  errors.New("unprocessable entity"),
	}
	serverErr = apiError{
		Status: http.StatusInternalServerError,
		Key:    "InternalServerErr",
		error:  errors.New("internal server error"),
	}
)

type apiError struct {
	// Status HTTP status code of the error.
	Status int `json:"status_code"`
	// Key returns a code that uniquely identifies an error.
	Key string `json:"key"`
	// Contains the API error message displayed to the user.
	Message string `json:"error"`
	// Contains the API error.
	error `json:"-"`
}

// Msg can be use to override the message error
// displayed to the user.
func (e apiError) Err(err error) apiError {
	e.error = err
	return e
}

func writeError(w http.ResponseWriter, req *http.Request, apiError apiError, err error) {
	var e *zerolog.Event
	if apiError.Status >= 500 {
		e = log.Error()
	} else {
		e = log.Warn()
	}

	e.Stack().
		Str("request_id", middleware.GetReqID(req.Context())).
		Str("url", req.URL.String()).
		Str("key", apiError.Key).
		Err(apiError).
		Err(err).
		Msg("API error")

	apiError.Message = apiError.Error()

	writeResponse(w, apiError.Status, apiError)
}

func writeResponse(w http.ResponseWriter, statusCode int, response interface{}) {
	w.WriteHeader(statusCode)

	if response == nil {
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("cannot marshal response")
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.Error().Err(err).Msg("cannot write response")
		return
	}
}

type LoginService interface {
	Login(ctx context.Context, login, password string) (service.User, error)
}

func (svc *ApiService) authMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				writeError(w, r, unauthorizedErr, fmt.Errorf("no valid basic auth found"))
				return
			}

			ctx := r.Context()
			u, err := svc.LoginService.Login(r.Context(), user, pass)
			if err != nil {
				switch errors.Unwrap(err) {
				case service.ErrUserNotFound:
					writeError(w, r, unauthorizedErr, err)
				default:
					writeError(w, r, forbiddenErr, err)
				}
				return
			}

			// Store the request user in the context.
			ctx = context.WithValue(ctx, userIDCtxKey{}, u)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func (svc *ApiService) requestIDResponseHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rID := middleware.GetReqID(r.Context())
			w.Header().Set(middleware.RequestIDHeader, rID)

			next.ServeHTTP(w, r)
		})
	}
}

func requestUser(r *http.Request) service.User {
	return r.Context().Value(userIDCtxKey{}).(service.User)
}

func urlParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// wrappedErrOrParent returns inner wrapped error if any.
func wrappedErrOrParent(err error) error {
	if wrappedErr := errors.Unwrap(err); wrappedErr != nil {
		return wrappedErrOrParent(wrappedErr)
	}

	return err
}
