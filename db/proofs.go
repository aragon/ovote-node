package db

import (
	"fmt"

	"github.com/aragon/zkmultisig-node/types"
)

// StoreProofID stores the given proofID for the given processID.  This method
// should be called only from a prover-server response.
func (r *SQLite) StoreProofID(processID, proofID uint64) error {
	sqlQuery := `
	INSERT INTO proofs(
		proofid,
		proof,
		insertedDatetime,
		processID
	) values(?, ?, CURRENT_TIMESTAMP, ?)
	`

	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	emptyProof := []byte{}
	_, err = stmt.Exec(proofID, emptyProof, processID)
	if err != nil {
		if err.Error() == "FOREIGN KEY constraint failed" {
			return fmt.Errorf("Can not store Proof, ProcessID=%d does not exist", processID)
		}
		return err
	}
	return nil
}

// AddProofToProofID stores the proof bytes for the given processID and
// proofID. Important: if the processID & proofID does not exist in the db yet,
// this method will not return any error, but will not store the data.
func (r *SQLite) AddProofToProofID(processID, proofID uint64, proof []byte) error {
	sqlQuery := `
	UPDATE proofs
	SET proof = ?
	WHERE (processID = ? AND proofID = ?)
	`

	stmt, err := r.db.Prepare(sqlQuery)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck

	_, err = stmt.Exec(proof, processID, proofID)
	if err != nil {
		return err
	}
	return nil
}

// GetProofsByProcessID returns the stored proofs for a given ProcessID
func (r *SQLite) GetProofsByProcessID(processID uint64) ([]types.ProofInDB, error) {
	rows, err := r.db.Query(
		"SELECT * FROM proofs WHERE processID = ?",
		processID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var proofs []types.ProofInDB
	for rows.Next() {
		proof := types.ProofInDB{}
		err = rows.Scan(&proof.ProofID, &proof.Proof, &proof.InsertedDatetime,
			&proof.ProcessID)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, proof)
	}
	return proofs, nil
}
