package http

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

type TransferRequest struct {
	DestinationWalletID *string  `json:"destination_wallet_id"`
	Amount              *float64 `json:"amount"`
	Message             *string  `json:"message"`
}

type TransferCreator interface {
	TransferFunds(ctx context.Context, req service.TransferRequest) (t service.Transfer, err error)
}

func CreateTransferHandler(svc TransferCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, r, readRequestBodyErr, err)
			return
		}

		var x TransferRequest
		if err = json.Unmarshal(b, &x); err != nil {
			writeError(w, r, unmarshalBodyErr, err)
			return
		}

		if x.DestinationWalletID == nil {
			writeError(w, r, badRequestErr.Err(errors.New("missing destination_wallet_id")), err)
			return
		}
		if x.Amount == nil {
			writeError(w, r, badRequestErr.Err(errors.New("missing amount")), err)
			return
		}
		if x.Message == nil {
			writeError(w, r, badRequestErr.Err(errors.New("missing message")), err)
			return
		}
		originWalletID, err := uuid.Parse(urlParam(r, "walletID"))
		if err != nil {
			writeError(w, r, notFoundErr, nil)
			return
		}
		destinationWalletID, err := uuid.Parse(*x.DestinationWalletID)
		if err != nil {
			writeError(w, r, badRequestErr.Err(errors.New("destination_id is not a valid user ID")), err)
			return
		}

		req := service.TransferRequest{
			Issuer:              requestUser(r),
			OriginWalletID:      originWalletID,
			DestinationWalletID: destinationWalletID,
			Amount:              *x.Amount,
			Message:             x.Message,
		}

		ctx := r.Context()
		t, err := svc.TransferFunds(ctx, req)
		if err != nil {
			switch wrappedErr := wrappedErrOrParent(err); wrappedErr {
			case service.ErrInsufficientFunds,
				service.ErrInvalidTransactionAmount,
				service.ErrSameTransferWallets:
				writeError(w, r, unprocessableEntityErr.Err(wrappedErr), err)
			case service.ErrWalletNotFound:
				writeError(w, r, notFoundErr.Err(wrappedErr), err)
			case service.ErrWalletAccessDenied:
				writeError(w, r, forbiddenErr.Err(wrappedErr), err)
			default:
				writeError(w, r, serverErr, err)
			}
			return
		}

		writeResponse(w, http.StatusCreated, t)
	}
}
