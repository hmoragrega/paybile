package http_test

import (
	"bytes"
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

func TestCreateTransferHandler(t *testing.T) {
	t.Parallel()
	var (
		user     = service.User{ID: uuid.MustParse("f4c34307-e7af-4add-a39b-b65d5627830c")}
		login    = "foo"
		pass     = "bar"
		auth     = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", login, pass)))
		headers  = map[string][]string{"Authorization": {"Basic " + auth}}
		message  = "transfer"
		transfer = service.Transfer{
			ID:                  uuid.MustParse("bc349396-fe96-42ae-ad10-d8e54d148c49"),
			IssuerID:            user.ID,
			OriginWalletID:      uuid.MustParse("b272dc21-e006-4a41-a120-2b8f26b61a67"),
			DestinationWalletID: uuid.MustParse("ba2428bd-88bb-44aa-9428-688d41817dc5"),
			Amount:              12.34,
			Message:             &message,
			Date:                time.Now(),
		}
	)

	transferJSON, _ := json.Marshal(&transfer)
	buildReq := func(method string, url *url.URL, headers map[string][]string, body string) http.Request {
		r, _ := http.NewRequest(method, "", bytes.NewBuffer([]byte(body)))
		r.Header = headers
		r.URL = url
		return *r
	}

	tt := []struct {
		name       string
		req        http.Request
		expect     func(cc *mocks.TransferCreator, ls *mocks.LoginService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "unauthorized",
			req: http.Request{
				Method: http.MethodPost,
				URL:    &url.URL{Path: "/api/v1/wallet/foo/transfer"},
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"status_code":401,"key":"UnauthorizedErr","error":"unauthorized"}`,
		},
		{
			name: "unmarshal error",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				"wrong",
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status_code":400,"key":"UnmarshalErr","error":"cannot unmarshal request body"}`,
		},
		{
			name: "missing destination wallet id",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"amount": 3, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status_code":400,"key":"BadRequestErr","error":"missing destination_wallet_id"}`,
		},
		{
			name: "missing amount",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "aeaf53d2-f238-4f6d-8ff9-f31b5dc3ec01", "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status_code":400,"key":"BadRequestErr","error":"missing amount"}`,
		},
		{
			name: "missing message",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "aeaf53d2-f238-4f6d-8ff9-f31b5dc3ec01", "amount": 3}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status_code":400,"key":"BadRequestErr","error":"missing message"}`,
		},
		{
			name: "invalid origin wallet id",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/foo/transfer"},
				headers,
				`{"destination_wallet_id": "aeaf53d2-f238-4f6d-8ff9-f31b5dc3ec01", "amount": 3, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"status_code":404,"key":"NotFoundErr","error":"not found"}`,
		},
		{
			name: "invalid destination wallet id",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "foo", "amount": 3, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"status_code":400,"key":"BadRequestErr","error":"destination_id is not a valid user ID"}`,
		},
		{
			name: "insufficient funds",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": 3, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, service.ErrInsufficientFunds)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"status_code":422,"key":"UnprocessableEntityErr","error":"origin has insufficient funds"}`,
		},
		{
			name: "invalid transaction amount",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": -1, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, service.ErrInvalidTransactionAmount)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"status_code":422,"key":"UnprocessableEntityErr","error":"invalid transaction amount"}`,
		},
		{
			name: "same wallet",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "b272dc21-e006-4a41-a120-2b8f26b61a67", "amount": 12.34, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, service.ErrSameTransferWallets)
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"status_code":422,"key":"UnprocessableEntityErr","error":"same origin and destination"}`,
		},
		{
			name: "wallet not found",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": 12.34, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, service.ErrWalletNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"status_code":404,"key":"NotFoundErr","error":"wallet not found"}`,
		},
		{
			name: "wallet access denied",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": 12.34, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, service.ErrWalletAccessDenied)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   `{"status_code":403,"key":"ForbiddenErr","error":"wallet access denied"}`,
		},
		{
			name: "internal server error",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": 12.34, "message": "foo"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(service.Transfer{}, fmt.Errorf("dummy"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"status_code":500,"key":"InternalServerErr","error":"internal server error"}`,
		},
		{
			name: "transfer ok",
			req: buildReq(
				http.MethodPost,
				&url.URL{Path: "/api/v1/wallet/b272dc21-e006-4a41-a120-2b8f26b61a67/transfer"},
				headers,
				`{"destination_wallet_id": "9c541caf-7185-456b-b418-5fa77cfbb687", "amount": 12.34, "message": "msg"}`,
			),
			expect: func(cc *mocks.TransferCreator, ls *mocks.LoginService) {
				ls.On("Login", mock.Anything, login, pass).Return(user, nil)
				cc.On("TransferFunds", mock.Anything, mock.Anything).Return(transfer, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   string(transferJSON),
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cc := &mocks.TransferCreator{}
			ls := &mocks.LoginService{}
			if tc.expect != nil {
				tc.expect(cc, ls)
			}

			api := httptransport.ApiService{
				TransferCreator: cc,
				LoginService:    ls,
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
