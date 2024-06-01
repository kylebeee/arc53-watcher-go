package misc

import "github.com/algorand/go-algorand-sdk/v2/types"

func ListInner(stxn *types.SignedTxnWithAD) []types.SignedTxnWithAD {
	txns := []types.SignedTxnWithAD{}
	for _, itxn := range stxn.ApplyData.EvalDelta.InnerTxns {
		txns = append(txns, itxn)
		txns = append(txns, ListInner(&itxn)...)
	}
	return txns
}
