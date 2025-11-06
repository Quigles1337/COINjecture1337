// Merkle tree utilities for computing tx_root and state_root
package consensus

import (
	"crypto/sha256"
)

// ComputeMerkleRoot computes the Merkle root of a list of hashes
// Uses a simple binary tree approach - institutional-grade implementation
func ComputeMerkleRoot(hashes [][32]byte) [32]byte {
	if len(hashes) == 0 {
		return [32]byte{} // Empty tree
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// Build Merkle tree bottom-up
	currentLevel := make([][32]byte, len(hashes))
	copy(currentLevel, hashes)

	for len(currentLevel) > 1 {
		var nextLevel [][32]byte

		// Process pairs
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				// Hash pair together
				combined := append(currentLevel[i][:], currentLevel[i+1][:]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash)
			} else {
				// Odd number of nodes - duplicate the last one
				combined := append(currentLevel[i][:], currentLevel[i][:]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash)
			}
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

// ComputeTxRoot computes the Merkle root of transaction hashes
func ComputeTxRoot(txHashes [][32]byte) [32]byte {
	return ComputeMerkleRoot(txHashes)
}

// ComputeStateRoot computes the Merkle root of account states
// For now, returns a placeholder. Will be implemented with full state tree later.
func ComputeStateRoot(accountHashes [][32]byte) [32]byte {
	return ComputeMerkleRoot(accountHashes)
}

// VerifyMerkleProof verifies a Merkle proof
// Returns true if the leaf is part of the tree with the given root
func VerifyMerkleProof(leaf [32]byte, proof [][32]byte, root [32]byte, index int) bool {
	currentHash := leaf

	for i, proofHash := range proof {
		// Determine if we should hash on left or right
		if (index>>i)&1 == 0 {
			// Left side
			combined := append(currentHash[:], proofHash[:]...)
			currentHash = sha256.Sum256(combined)
		} else {
			// Right side
			combined := append(proofHash[:], currentHash[:]...)
			currentHash = sha256.Sum256(combined)
		}
	}

	return currentHash == root
}
