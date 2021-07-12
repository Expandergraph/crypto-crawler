package collector

import (
	"context"
	"math/big"
	"time"

	"github.com/Expandergraph/crypto-crawler/database"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Expandergraph/crypto-crawler/model"
	"github.com/Expandergraph/crypto-crawler/rpc"
)

const (
	DefaultPollingInterval = 10 * time.Second
	module                 = "collector"
)

type Manager struct {
	ctx context.Context

	requester    *rpc.ETHRPCRequester
	db           *database.DB
	currentBlock uint64
}

func New(ctx context.Context, db *database.DB, url string) *Manager {
	requester := rpc.NewETHRPCRequester(url)

	return &Manager{
		ctx:       ctx,
		db:        db,
		requester: requester,
	}
}

func (m *Manager) Run() error {
	ticker := time.NewTicker(DefaultPollingInterval)
	for {
		select {
		case <-m.ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			if err := m.SyncToLatestBlock(); err != nil {
				log.WithFields(log.Fields{"module": module, "err": err}).Error("failed on sync to latest block")
			}
		}
	}
}

func (m *Manager) SyncToLatestBlock() error {
	latestNumber, err := m.requester.GetLatestBlockNumber()
	if err != nil {
		return errors.Wrap(err, "failed on get latest block number")
	}

	syncedBlock, err := m.db.QuerySyncedBlock()
	if err != nil {
		return errors.Wrap(err, "failed on get synced block number")
	}

	for i := syncedBlock; i <= latestNumber.Uint64(); i++ {
		block, logs, err := m.ProcessBlockByNumber(int64(i))
		if err != nil {
			return errors.Wrap(err, "failed on get block by number")
		}

		//write to db
		if err := m.db.UpdateSyncedBlock(block, logs); err != nil {
			return errors.Wrapf(err, "failed update block %d to db", i)
		}
		log.WithFields(log.Fields{"module": module, "block": i}).Info("sync block successfully")
	}

	return nil
}

func (m *Manager) ProcessBlockByNumber(number int64) (*model.FullBlock, []model.Log, error) {
	fullBlock, err := m.requester.GetBlockInfoByNumber(big.NewInt(number))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed on get full block info")
	}

	logs, err := m.requester.EthGetLogs(model.FilterParams{
		FromBlock: fullBlock.Number,
		ToBlock:   fullBlock.Number,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed on get logs")
	}

	return fullBlock, logs, nil
}
