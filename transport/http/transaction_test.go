package http_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
	httptransport "github.com/hmoragrega/paybile/transport/http"
	"github.com/hmoragrega/paybile/transport/http/mocks"
	"github.com/stretchr/testify/mock"
)

func TestTransactionListHandler(t *testing.T) {
	t.Parallel()
	var (
		user     = service.User{ID: uuid.MustParse("f4c34307-e7af-4add-a39b-b65d5627830c")}
		login    = "foo"
		pass     = "bar"
		auth     = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", login, pass)))
		headers  = map[string][]string{"Authorization": {"Basic " + auth}}
		walletID = uuid.MustParse("f8ee8d07-a56a-49f1-87ee-b4377c8142bb")
		fromID   = uuid.MustParse("fe2cd404-8aee-4c25-be3a-15f9354f717a")
		list     = service.TransactionList{
			Results: []service.Transaction{{
				ID:      uuid.MustParse("53237090-0d16-4447-93df-394df1e4c7c8"),
				Amount:  10,
				Balance: 20,
				Date:    time.Now(),
			}},
		}
	)
	listBody, _ := json.Marshal(&list)

	tt := []struct {
		name       string
		req        http.Request
		expect     func(tl *mocks.TransactionLister, ls *mocks.LoginService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "unauthorized",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/foo/transactions"},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"status_code":401,"key":"UnauthorizedErr","error":"unauthorized"}`,
		},
		{
			name: "invalid origin wallet id",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/foo/transactions"},
				Header: headers,
			},
			expect: func(tl *mocks.TransactionLister, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"status_code":404,"key":"NotFoundErr","error":"not found"}`,
		},
		{
			name: "invalid from id",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/f8ee8d07-a56a-49f1-87ee-b4377c8142bb/transactions", RawQuery: "from_id=foo"},
				Header: headers,
			},
			expect: func(tl *mocks.TransactionLister, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"status_code":422,"key":"UnprocessableEntityErr","error":"from_id is not a valid UUID"}`,
		},
		{
			name: "list error",
			req: http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: "/api/v1/wallet/f8ee8d07-a56a-49f1-87ee-b4377c8142bb/transactions"},
				Header: headers,
			},
			expect: func(tl *mocks.TransactionLister, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				tl.On("ListTransactions", mock.Anything, user, walletID, mock.Anything).Return(list, fmt.Errorf("dummy"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"status_code":500,"key":"InternalServerErr","error":"internal server error"}`,
		},
		{
			name: "list ok",
			req: http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Path:     "/api/v1/wallet/f8ee8d07-a56a-49f1-87ee-b4377c8142bb/transactions",
					RawQuery: "from_id=" + fromID.String() + "&per_page=10&order=desc",
				},
				Header: headers,
			},
			expect: func(tl *mocks.TransactionLister, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				tl.On("ListTransactions", mock.Anything, user, walletID, service.ListOptions{
					FromID:  &fromID,
					Order:   service.Descending,
					PerPage: 10,
				}).Return(list, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   string(listBody),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tl := &mocks.TransactionLister{}
			ls := &mocks.LoginService{}
			if tc.expect != nil {
				tc.expect(tl, ls)
			}

			api := httptransport.ApiService{
				TransactionLister: tl,
				LoginService:      ls,
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
