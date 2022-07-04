//go:build unit

package accounts

import (
	"context"
	"database/sql"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)
	holderRepoMock := holders.NewMockRepository(ctrl)

	svc := NewService(tracer.NewNoop(), repoMock, holderRepoMock)

	t.Run("fail create, holder not found", func(t *testing.T) {
		account := Account{
			Name:           gofakeit.Name(),
			DocumentNumber: gofakeit.SSN(),
		}
		holderRepoMock.EXPECT().
			GetByFilter(ctx, holders.HolderFilter{DocumentNumber: account.DocumentNumber}).
			Return([]holders.HolderModel{}, sql.ErrNoRows)

		created, err := svc.Create(ctx, account)
		assert.EqualError(t, err, "sql: no rows in result set")
		assert.Empty(t, created)
	})

	t.Run("success create", func(t *testing.T) {
		account := Account{
			Name:           gofakeit.Name(),
			DocumentNumber: gofakeit.SSN(),
		}

		holderRepoMock.EXPECT().
			GetByFilter(ctx, holders.HolderFilter{DocumentNumber: account.DocumentNumber}).
			Return(
				[]holders.HolderModel{
					{
						ID:             uuid.New(),
						Name:           gofakeit.Name(),
						DocumentNumber: account.DocumentNumber,
					},
				},
				nil,
			)
		repoMock.EXPECT().
			Create(ctx, gomock.Any()).
			Return(
				accountModel{
					ID:       uuid.New(),
					Name:     account.Name,
					Agency:   "0001",
					Number:   "123120",
					HolderID: account.HolderID,
					Status:   ActiveStatus,
				}, nil,
			)

		created, err := svc.Create(ctx, account)
		assert.NoError(t, err)
		assert.NotEmpty(t, created)
	})
}

func TestService_BlockByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)

	svc := NewService(tracer.NewNoop(), repoMock, holders.NewMockRepository(ctrl))

	accountID := uuid.New()

	t.Run("fail block, not found", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{}, sql.ErrNoRows)

		acc, err := svc.BlockByID(ctx, accountID)
		assert.EqualError(t, err, "sql: no rows in result set")
		assert.Empty(t, acc)
	})

	t.Run("fail block, not active", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: ClosedStatus},
			}, nil)

		acc, err := svc.BlockByID(ctx, accountID)
		assert.EqualError(t, err, "account must be active for this operation")
		assert.Empty(t, acc)
	})

	t.Run("success block", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: ActiveStatus},
			}, nil)

		repoMock.EXPECT().
			Update(ctx, accountModel{Status: BlockedStatus}).
			Return(accountModel{Status: BlockedStatus}, nil)

		acc, err := svc.BlockByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, BlockedStatus, acc.Status)
	})
}

func TestService_UnblockByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)

	svc := NewService(tracer.NewNoop(), repoMock, holders.NewMockRepository(ctrl))

	accountID := uuid.New()

	t.Run("fail unblock, not found", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{}, sql.ErrNoRows)

		acc, err := svc.UnblockByID(ctx, accountID)
		assert.EqualError(t, err, "sql: no rows in result set")
		assert.Empty(t, acc)
	})

	t.Run("fail unblock, not blocked", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: ActiveStatus},
			}, nil)

		acc, err := svc.UnblockByID(ctx, accountID)
		assert.EqualError(t, err, "account must be blocked for this operation")
		assert.Empty(t, acc)
	})

	t.Run("success unblock", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: BlockedStatus},
			}, nil)

		repoMock.EXPECT().
			Update(ctx, accountModel{Status: ActiveStatus}).
			Return(accountModel{Status: ActiveStatus}, nil)

		acc, err := svc.UnblockByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, ActiveStatus, acc.Status)
	})
}

func TestService_CloseByID(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := NewMockRepository(ctrl)

	svc := NewService(tracer.NewNoop(), repoMock, holders.NewMockRepository(ctrl))

	accountID := uuid.New()

	t.Run("fail close, not found", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{}, sql.ErrNoRows)

		acc, err := svc.CloseByID(ctx, accountID)
		assert.EqualError(t, err, "sql: no rows in result set")
		assert.Empty(t, acc)
	})

	t.Run("success close", func(t *testing.T) {
		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: ActiveStatus},
			}, nil)

		repoMock.EXPECT().
			Update(ctx, accountModel{Status: ClosedStatus}).
			Return(accountModel{Status: ClosedStatus}, nil)

		acc, err := svc.CloseByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, ClosedStatus, acc.Status)

		repoMock.EXPECT().
			GetByFilter(ctx, accountFilter{ID: uuid.NullUUID{UUID: accountID, Valid: true}}).
			Return([]accountModel{
				{Status: ClosedStatus},
			}, nil)

		acc, err = svc.CloseByID(ctx, accountID)
		assert.NoError(t, err)
		assert.Equal(t, ClosedStatus, acc.Status)
	})
}
