package db

import (
	"database/sql"

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
		processID INTEGER NOT NULL PRIMARY KEY UNIQUE,
		censusRoot BLOB NOT NULL,
		ethBlockNum INTEGER NOT NULL,
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
		FOREIGN KEY(processID) REFERENCES processes(processID)
	);
	`
	_, err = r.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// StoreProcess stores a new process with the given id, censusRoot and
// ethBlockNum
func (r *SQLite) StoreProcess(id uint64, censusRoot []byte, ethBlockNum uint64) error {
	sqlAddvote := `
	INSERT INTO processes(
		processID,
		censusRoot,
		ethBlockNum,
		insertedDatetime
	) values(?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := r.db.Prepare(sqlAddvote)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(id, censusRoot, ethBlockNum)
	if err != nil {
		return err
	}
	return nil
}

// ReadProcesses reads all the stored types.Process
func (r *SQLite) ReadProcesses() ([]types.Process, error) {
	sqlReadall := `
	SELECT processID, censusRoot, ethBlockNum, insertedDatetime FROM processes
	ORDER BY datetime(InsertedDatetime) DESC
	`

	rows, err := r.db.Query(sqlReadall)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var processes []types.Process
	for rows.Next() {
		process := types.Process{}
		err = rows.Scan(&process.ID, &process.CensusRoot,
			&process.EthBlockNum, &process.InsertedDatetime)
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
		return err
	}
	return nil
}

// ReadVotePackagesByProcessID reads all the stored types.VotePackage for the
// given ProcessID
func (r *SQLite) ReadVotePackagesByProcessID(processID uint64) ([]types.VotePackage, error) {
	// TODO add pagination
	sqlReadall := `
	SELECT signature, indx, publicKey, merkleproof, vote FROM votepackages
	WHERE processID = ?
	ORDER BY datetime(InsertedDatetime) DESC
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
