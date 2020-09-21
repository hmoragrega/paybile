package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
)

type WalletGetter interface {
	GetWallet(ctx context.Context, user service.User, walletID uuid.UUID) (service.Wallet, error)
}

func GetWalletHandler(svc WalletGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		walletID, err := uuid.Parse(urlParam(r, "walletID"))
		if err != nil {
			writeError(w, r, notFoundErr, nil)
			return
		}

		x, err := svc.GetWallet(r.Context(), requestUser(r), walletID)
		if err != nil {
			switch wrappedErr := wrappedErrOrParent(err); wrappedErr {
			case service.ErrWalletNotFound:
				writeError(w, r, notFoundErr.Err(wrappedErr), err)
			case service.ErrWalletAccessDenied:
				writeError(w, r, forbiddenErr.Err(wrappedErr), err)
			default:
				writeError(w, r, serverErr, err)
			}
			return
		}

		writeResponse(w, http.StatusOK, x)
	}
}
