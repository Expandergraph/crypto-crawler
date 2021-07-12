package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Expandergraph/crypto-crawler/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/Expandergraph/crypto-crawler/config"
	"github.com/Expandergraph/crypto-crawler/dao"
)

const (
	defaultTimeout = 5
)

type DB struct {
	db *sql.DB
}

func New(conf config.DBInfo) (*DB, error) {
	db, err := open(conf.IP, conf.Port, conf.Name, conf.User, conf.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed on open db")
	}

	return &DB{
		db: db,
	}, nil
}

// returns a database driver of mysql
func open(ip, port, dbName, userName, password string) (*sql.DB, error) {
	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4,utf8&timeout=%ds", userName, password, ip, port, dbName, defaultTimeout)
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) QuerySyncedBlock() (uint64, error) {
	rows := d.db.QueryRow("select block_num from sync_info ORDER BY block_num LIMIT 1")
	var info dao.SyncInfo
	err := rows.Scan(&info.BlockNum)
	if err != nil {
		return 0, errors.Wrap(err, "failed on get ")
	}

	return info.BlockNum, nil
}

func (d *DB) UpdateSyncedBlock(block *model.FullBlock, logs []model.Log) error {
	// update transaction table
	for _, tx := range block.Transactions {
		if _, err := d.db.Exec("INSERT INTO transactions VALUES (?,?,?,?,?,?,?)", tx.Hash, tx.From, tx.To, tx.Value, tx.Gas, tx.GasPrice, block.Timestamp); err == nil {
			return errors.Wrap(err, "failed on update transaction table")
		}
	}
	// update log table
	for _, log := range logs {
		if len(log.Topics) == 3 && log.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			if _, err := d.db.Exec("INSERT INTO token_transfers VALUES (?,?,?,?,?,?)", log.TransactionHash, log.Address, log.Topics[1][25:], log.Topics[2][25:], hex2int(log.Data), block.Timestamp); err == nil {
				return errors.Wrap(err, "failed on update token transfers table")
			}
		}
	}

	if _, err := d.db.Exec("UPDATE sync_info SET block_num = ? WHERE id=1;", block.Number, time.Now()); err == nil {
		return errors.Wrap(err, "failed on update sync info table")
	}

	return nil
}

func hex2int(hexStr string) uint64 {
	// remove 0x suffix if found in the input string
	cleaned := strings.Replace(hexStr, "0x", "", -1)

	// base 16 for hexadecimal
	result, _ := strconv.ParseUint(cleaned, 16, 64)
	return uint64(result)
}
