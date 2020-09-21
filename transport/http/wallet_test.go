package http_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
	httptransport "github.com/hmoragrega/paybile/transport/http"
	"github.com/hmoragrega/paybile/transport/http/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	m.Run()
}

func TestGetWalletHandler(t *testing.T) {
	t.Parallel()
	var (
		user     = service.User{ID: uuid.MustParse("f4c34307-e7af-4add-a39b-b65d5627830c")}
		walletID = uuid.MustParse("b272dc21-e006-4a41-a120-2b8f26b61a67")
		login    = "foo"
		pass     = "bar"
		auth     = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", login, pass)))
		wallet   = service.Wallet{
			ID:      walletID,
			UserID:  user.ID,
			Balance: 10.99,
		}
		headers = map[string][]string{
			"Authorization": {"Basic " + auth},
		}
	)
	tt := []struct {
		name       string
		req        http.Request
		expect     func(wg *mocks.WalletGetter, ls *mocks.LoginService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "unauthorized",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/foo"},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"status_code":401,"key":"UnauthorizedErr","error":"unauthorized"}`,
		},
		{
			name: "invalid wallet id",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/foo"},
				Header: headers,
			},
			expect: func(wg *mocks.WalletGetter, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"status_code":404,"key":"NotFoundErr","error":"not found"}`,
		},
		{
			name: "wallet not found",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/" + walletID.String()},
				Header: headers,
			},
			expect: func(wg *mocks.WalletGetter, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				wg.On("GetWallet", mock.Anything, user, walletID).Return(service.Wallet{}, service.ErrWalletNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"status_code":404,"key":"NotFoundErr","error":"wallet not found"}`,
		},
		{
			name: "wallet access denied",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/" + walletID.String()},
				Header: headers,
			},
			expect: func(wg *mocks.WalletGetter, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				wg.On("GetWallet", mock.Anything, user, walletID).Return(service.Wallet{}, service.ErrWalletAccessDenied)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `{"status_code":403,"key":"ForbiddenErr","error":"wallet access denied"}`,
		},
		{
			name: "server error",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/" + walletID.String()},
				Header: headers,
			},
			expect: func(wg *mocks.WalletGetter, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				wg.On("GetWallet", mock.Anything, user, walletID).Return(service.Wallet{}, fmt.Errorf("dummy"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"status_code":500,"key":"InternalServerErr","error":"internal server error"}`,
		},
		{
			name: "ok",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/" + walletID.String()},
				Header: headers,
			},
			expect: func(wg *mocks.WalletGetter, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				wg.On("GetWallet", mock.Anything, user, walletID).Return(wallet, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"id":"b272dc21-e006-4a41-a120-2b8f26b61a67","user_id":"f4c34307-e7af-4add-a39b-b65d5627830c","balance":10.99}`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			wg := &mocks.WalletGetter{}
			ls := &mocks.LoginService{}
			if tc.expect != nil {
				tc.expect(wg, ls)
			}

			api := httptransport.ApiService{
				WalletGetter: wg,
				LoginService: ls,
			}
			rr := httptest.NewRecorder()
			api.Handler().ServeHTTP(rr, &tc.req)
			res := rr.Result()

			if got := res.StatusCode; got != tc.wantStatus {
				t.Fatalf("unexpected status: \n got:  %+v \n want: %+v", got, tc.wantStatus)
			}
			if got := rr.Body.String(); got != tc.wantBody {
				t.Fatalf("unexpected body: \n got:  %+v \n want: %+v", got, tc.wantBody)
			}
		})
	}
}
