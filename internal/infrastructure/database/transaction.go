package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// contextKey コンテキストのキーの型
type contextKey string

const (
	// txKey トランザクションをコンテキストに保存するためのキー
	txKey contextKey = "tx"
)

// TransactionManager トランザクション管理のインターフェース
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TxGetter トランザクションを取得するインターフェース
type TxGetter interface {
	GetTx(ctx context.Context) (*sqlx.Tx, bool)
}

// Executor クエリを実行するためのインターフェース
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// txManager TransactionManagerの実装
type txManager struct {
	db *sqlx.DB
}

// NewTransactionManager 新しいTransactionManagerを作成
func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &txManager{
		db: db,
	}
}

// RunInTransaction トランザクション内で関数を実行
func (tm *txManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 既存のトランザクションがある場合はそれを使用
	if _, ok := GetTx(ctx); ok {
		return fn(ctx)
	}

	// 新しいトランザクションを開始
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// トランザクションをコンテキストに保存
	ctx = WithTx(ctx, tx)

	// deferでロールバックまたはコミット
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // ロールバック後にpanicを再スロー
		}
	}()

	// 関数を実行
	if err := fn(ctx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// コミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTx コンテキストにトランザクションを設定
func WithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTx コンテキストからトランザクションを取得
func GetTx(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sqlx.Tx)
	return tx, ok
}

// GetExecutor コンテキストから適切なExecutorを取得
// トランザクションがあればそれを、なければDBを返す
func GetExecutor(ctx context.Context, db *sqlx.DB) Executor {
	if tx, ok := GetTx(ctx); ok {
		return tx
	}
	return db
}

// TxOptions トランザクションのオプション
type TxOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}
