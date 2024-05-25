package atomiccommit

import (
	"context"
)

type Status int

const (
	Active    Status = iota //Transaction in progress
	Commiting               //Transaction commit started
	Commited                //Transaction commit finished successfull
	Aborting                //Transaction abprt stated
	Aborted                 //Transaction was aborted by user
)

// Transaction represents a transaction.
// ... and should be completed by user via either Commit or Abort.

type Transaction interface {
	User() string        //User name associated  with transaction
	Description() string //description of transaction

	Extention() string

	Status() Status

	//commit finalizes the transaction
	Commit(ctx context.Context) error

	//abort completes the transaction by executing Abort on All
	Abort()

	//join associated data managers will participate in the transaction
	Join(dm DataManager)

	//RegisterSync register sync to be notified of this transaction boundary events
	RegisterSync(sync Synchoronizer)
}

// Datamanger manages data and can transactionally persist it .
type DataManager interface {
	//ABort should abort all modification to managed data
	Abort(txn Transaction)

	//TCPBegin should begin commit of a transaction, starting the two-phase-commit
	TCPBegin(txn Transaction)

	//Commit should begin commit of a tranaction, starting the two-phase-commit
	Commit(ctx context.Context, txn Transaction) error

	//TCPVote should verify that a data manager can commit the transaction.
	TCPVote(ctx context.Context, txn Transaction) error

	//TCPFinish should indicate confirmation that the transaction is done.
	TCPFinish(cyx context.Context, txn Transaction)
}

// Synchromizeer is the interface layer to participate in transaction-boundary notification.
type Synchoronizer interface {
	//Before Completion is called before corresponding transaction is going to be completed.
	BeforeCompletion(txn Transaction)
	AfterCompletion(txn Transaction)
}

// func New craete New Transaction
func New(ctx context.Context) Transaction {
	return New(ctx)
}

// Current returns current transaction.
func Current(ctx context.Context) Transaction {
	return currentTxn(ctx)
}

// With runs f in a new transaction, and either commits or aborts it depending on f result.
func With(ctx context.Context, f func(context.Context) error) (ok bool, _ error) {
	txn, ctx := newTxn(ctx)
	err := f(ctx)
	if err != nil {
		txn.Abort() //.err
		return false, err
	}
	return true, txn.Commit(ctx)
}
