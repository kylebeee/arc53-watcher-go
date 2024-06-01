package server

import (
	"fmt"
	"strings"

	"github.com/kylebeee/arc53-watcher-go/internal/algod"
	"github.com/kylebeee/arc53-watcher-go/misc"
)

func (s *Arc53WatcherServer) ProcessBlock(b *algod.BlockWrap) {
	fmt.Printf("\n\n[BLK]: %v\n", b.Block.Round)

	for i := range b.Block.Payset {
		stxn := b.Block.Payset[i]
		txn := b.Block.Payset[i].SignedTxnWithAD.SignedTxn.Txn

		id, err := algod.DecodeTxnId(b.Block.BlockHeader, &stxn)
		if err != nil {
			fmt.Println(err)
			s.ProcessingFailures = append(s.ProcessingFailures, err)
			continue
		}

		if s.PrintTxns {
			fmt.Printf("[TXN]%s[%s]: %s\n", strings.Repeat(" ", 6-len(string(txn.Type))), strings.ToUpper(string(txn.Type)), id)
			innerTxns := misc.ListInner(&stxn.SignedTxnWithAD)
			if len(innerTxns) > 0 {
				for i := range innerTxns {
					stxn := innerTxns[i]
					fmt.Printf("     %s[%s]: [%v]\n", strings.Repeat(" ", 6-len(string(stxn.Txn.Type))), strings.ToUpper(string(stxn.Txn.Type)), i)
				}
			}
		}

		for i := range s.Providers {
			err := s.Providers[i].Process(stxn, uint64(b.Block.Round))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
