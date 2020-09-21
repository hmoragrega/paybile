package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

type TransactionLister interface {
	ListTransactions(ctx context.Context, user service.User, walletID uuid.UUID, opt service.ListOptions) (l service.TransactionList, err error)
}

func TransactionListHandler(svc TransactionLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			order   = r.URL.Query().Get("order")
			perPage = r.URL.Query().Get("per_page")
			fromID  = r.URL.Query().Get("from_id")
			user    = requestUser(r)
		)

		walletID, err := uuid.Parse(urlParam(r, "walletID"))
		if err != nil {
			writeError(w, r, notFoundErr, nil)
			return
		}

		opts := service.ListOptions{
			Order:   service.ParseListOrderOrDefault(order),
			PerPage: service.ParsePerPageOrDefault(perPage),
		}

		if fromID != "" {
			if x, err := uuid.Parse(fromID); err != nil {
				writeError(w, r, unprocessableEntityErr.Err(errors.New("from_id is not a valid UUID")), err)
				return
			} else {
				opts.FromID = &x
			}
		}

		l, err := svc.ListTransactions(r.Context(), user, walletID, opts)
		if err != nil {
			writeError(w, r, serverErr, err)
			return
		}

		writeResponse(w, http.StatusOK, l)
	}
}
