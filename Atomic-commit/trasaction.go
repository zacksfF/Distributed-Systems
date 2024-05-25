package atomiccommit

import (
	"context"
	"sync"
)

// Package transaction provides transaction management via two-phase commit protocol.

// transaction implements Transaction
type transaction struct {
	mu     sync.Mutex
	status Status
	datav  []DataManager
	syncv  []Synchoronizer

	//metadata
	user        string
	description string
	extention   string
}

type CtxKey struct{}

// getTxn returns transaction associated with provided context.
// nil is returned if there is no association.
func getTxn(ctx context.Context) *transaction {
	t := ctx.Value(CtxKey{})
	if t == nil {
		return nil
	}
	return t.(*transaction)
}

// currentTxn serves Current.
func currentTxn(ctx context.Context) Transaction {
	txn := getTxn(ctx)
	if txn == nil {
		panic("transaction: no current transaction")
	}
	return txn
}

// newTxn serves New.
func newTxn(ctx context.Context) (Transaction, context.Context) {
	if getTxn(ctx) != nil {
		panic("transactiom : new: nested transaction not supported")
	}

	txn := &transaction{
		status: Active,
	}
	txnCtx := context.WithValue(ctx, CtxKey{}, txn)
	return txn, txnCtx
}

// Status implements Transaction.
func (txn *transaction) Status() Status {
	txn.mu.Lock()
	defer txn.mu.Unlock()
	return txn.status
}

// commit implemts Transaction.
func (txn *transaction) Commit(ctx context.Context) error {
	panic("TODO")
}



// Join implements Transaction.
func (txn *transaction) Join(dm DataManager) {
	txn.mu.Lock()
	defer txn.mu.Unlock()

	txn.checkNotYetCompleting("join")

	// XXX forbid double join?
	txn.datav = append(txn.datav, dm)
}

// RegisterSync implements Transaction.
func (txn *transaction) RegisterSync(sync Synchoronizer) {
	txn.mu.Lock()
	defer txn.mu.Unlock()

	txn.checkNotYetCompleting("register sync")

	// XXX forbid double register?
	txn.syncv = append(txn.syncv, sync)
}

// checkNotYetCompleting asserts that transaction completion has not yet began.
//
// and panics if the assert fails.
// must be called with .mu held.
func (txn *transaction) checkNotYetCompleting(who string) {
	switch txn.status {
	case Active: // XXX + Doomed ?
		// ok
	default:
		panic("transaction: " + who + ": transaction completion already began")
	}
}

// ---- meta ----

func (txn *transaction) User() string        { return txn.user }
func (txn *transaction) Description() string { return txn.description }
func (txn *transaction) Extension() string   { return txn.extention }
