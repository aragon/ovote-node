package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/aragon/zkmultisig-node/types"
)

// SQLite represents the SQLite database
type SQLite struct {
	db *sql.DB
}

// NewSQLite returns a new *SQLite database
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{
		db: db,
	}
}

// Migrate creates the tables needed for the database
func (r *SQLite) Migrate() error {
	query := `
	PRAGMA foreign_keys = ON;
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}

	query = `
	CREATE TABLE IF NOT EXISTS processes(
		id INTEGER NOT NULL PRIMARY KEY UNIQUE,
		status INTEGER NOT NULL,
		censusRoot BLOB NOT NULL,
		ethBlockNum INTEGER NOT NULL,
		ethEndBlockNum INTEGER NOT NULL,
		insertedDatetime DATETIME
	);
	`
	_, err = r.db.Exec(query)
	if err != nil {
		return err
	}

	query = `
	CREATE TABLE IF NOT EXISTS votepackages(
		indx INTEGER NOT NULL PRIMARY KEY UNIQUE,
		publicKey BLOB NOT NULL UNIQUE,
		merkleproof BLOB NOT NULL UNIQUE,
		signature BLOB NOT NULL,
		vote BLOB NOT NULL,
		insertedDatetime DATETIME,
		processID INTEGER NOT NULL,
		FOREIGN KEY(processID) REFERENCES processes(id)
	);
	`
	_, err = r.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// StoreProcess stores a new process with the given id, censusRoot and
// ethBlockNum. When a new process is stored, it's assumed that it comes from
// the SmartContract, and its status is set to types.ProcessStatusOn
func (r *SQLite) StoreProcess(id uint64, censusRoot []byte, ethBlockNum,
	ethEndBlockNum uint64) error {
	sqlAddvote := `
	INSERT INTO processes(
		id,
		status,
		censusRoot,
		ethBlockNum,
		ethEndBlockNum,
		insertedDatetime
	) values(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := r.db.Prepare(sqlAddvote)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(id, types.ProcessStatusOn, censusRoot, ethBlockNum, ethEndBlockNum)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProcessStatus sets the given types.ProcessStatus for the given id.
// This method should only be called when updating from SmartContracts.
func (r *SQLite) UpdateProcessStatus(id uint64, status types.ProcessStatus) error {
	sqlAddvote := `
	UPDATE processes SET status=? WHERE id=?
	`

	stmt, err := r.db.Prepare(sqlAddvote)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(int(status), id)
	if err != nil {
		return err
	}
	return nil
}

// GetProcessStatus returns the stored types.ProcessStatus for the given id
func (r *SQLite) GetProcessStatus(id uint64) (types.ProcessStatus, error) {
	row := r.db.QueryRow("SELECT status FROM processes WHERE id = ?", id)

	var status int
	err := row.Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("Process ID:%d, does not exist in the db", id)
		}
		return 0, err
	}
	return types.ProcessStatus(status), nil
}

// ReadProcessByID reads the types.Process by the given id
func (r *SQLite) ReadProcessByID(id uint64) (*types.Process, error) {
	row := r.db.QueryRow("SELECT * FROM processes WHERE id = ?", id)

	var process types.Process
	err := row.Scan(&process.ID, &process.Status, &process.CensusRoot,
		&process.EthBlockNum, &process.EthEndBlockNum, &process.InsertedDatetime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("Process ID:%d, does not exist in the db", id)
		}
		return nil, err
	}
	return &process, nil
}

// ReadProcesses reads all the stored types.Process
func (r *SQLite) ReadProcesses() ([]types.Process, error) {
	sqlReadall := `
	SELECT * FROM processes	ORDER BY datetime(insertedDatetime) DESC
	`
	// TODO maybe, in all affected methods, order by EthBlockNum (creation)
	// instead of insertedDatetime.

	rows, err := r.db.Query(sqlReadall)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var processes []types.Process
	for rows.Next() {
		process := types.Process{}
		err = rows.Scan(&process.ID, &process.Status,
			&process.CensusRoot, &process.EthBlockNum,
			&process.EthEndBlockNum, &process.InsertedDatetime)
		if err != nil {
			return nil, err
		}
		processes = append(processes, process)
	}
	return processes, nil
}

// ReadProcessesByEthEndBlockNum reads all the stored processes which contain
// the given EthEndBlockNum
func (r *SQLite) ReadProcessesByEthEndBlockNum(ethEndBlockNum uint64) ([]types.Process, error) {
	sqlReadall := `
	SELECT * FROM processes WHERE ethEndBlockNum = ?
	ORDER BY datetime(insertedDatetime) DESC
	`

	rows, err := r.db.Query(sqlReadall, ethEndBlockNum)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var processes []types.Process
	for rows.Next() {
		process := types.Process{}
		err = rows.Scan(&process.ID, &process.Status,
			&process.CensusRoot, &process.EthBlockNum,
			&process.EthEndBlockNum, &process.InsertedDatetime)
		if err != nil {
			return nil, err
		}
		processes = append(processes, process)
	}
	return processes, nil
}

// ReadProcessesByStatus reads all the stored processes which have the given
// status
func (r *SQLite) ReadProcessesByStatus(status types.ProcessStatus) ([]types.Process, error) {
	sqlReadall := `
	SELECT * FROM processes WHERE status = ?
	ORDER BY datetime(insertedDatetime) DESC
	`

	rows, err := r.db.Query(sqlReadall, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var processes []types.Process
	for rows.Next() {
		process := types.Process{}
		err = rows.Scan(&process.ID, &process.Status,
			&process.CensusRoot, &process.EthBlockNum,
			&process.EthEndBlockNum, &process.InsertedDatetime)
		if err != nil {
			return nil, err
		}
		processes = append(processes, process)
	}
	return processes, nil
}

// StoreVotePackage stores the given types.VotePackage for the given CensusRoot
func (r *SQLite) StoreVotePackage(processID uint64, vote types.VotePackage) error {
	// TODO check that processID exists
	sqlAddvote := `
	INSERT INTO votepackages(
		indx,
		publicKey,
		merkleproof,
		signature,
		vote,
		insertedDatetime,
		processID
	) values(?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?)
	`

	stmt, err := r.db.Prepare(sqlAddvote)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(vote.CensusProof.Index, vote.CensusProof.PublicKey,
		vote.CensusProof.MerkleProof, vote.Signature[:], vote.Vote, processID)
	if err != nil {
		if err.Error() == "FOREIGN KEY constraint failed" {
			return fmt.Errorf("Can not store VotePackage, ProcessID=%d does not exist", processID)
		}
		return err
	}
	return nil
}

// ReadVotePackagesByProcessID reads all the stored types.VotePackage for the
// given ProcessID. VotePackages returned are sorted by index parameter, from
// smaller to bigger.
func (r *SQLite) ReadVotePackagesByProcessID(processID uint64) ([]types.VotePackage, error) {
	// TODO add pagination
	sqlReadall := `
	SELECT signature, indx, publicKey, merkleproof, vote FROM votepackages
	WHERE processID = ?
	ORDER BY datetime(indx) DESC
	`

	rows, err := r.db.Query(sqlReadall, processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var votes []types.VotePackage
	for rows.Next() {
		vote := types.VotePackage{}
		var sigBytes []byte
		err = rows.Scan(&sigBytes, &vote.CensusProof.Index,
			&vote.CensusProof.PublicKey, &vote.CensusProof.MerkleProof,
			&vote.Vote)
		if err != nil {
			return nil, err
		}
		copy(vote.Signature[:], sigBytes)
		votes = append(votes, vote)
	}
	return votes, nil
}

// func (r *SQLite) ReadVotePackagesByCensusRoot(processID uint64) ([]types.VotePackage, error) {
// func (r *SQLite) ReadVoteByPublicKeyAndCensusRoot(censusRoot []byte) (
// 	[]types.VotePackage, error) {
// func (r *SQLite) ReadVotesByPublicKey(censusRoot []byte) ([]types.VotePackage, error) {
