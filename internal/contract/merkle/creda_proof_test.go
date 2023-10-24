package bloom

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"golang.org/x/crypto/sha3"
)

func keccak256(s1, s2 string) string {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(s1 + s2))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func buildMerkleTreeAndProof(addresses []string, scores []string, targetAddress string) (string, []string) {
	leaves := make([]string, len(addresses))
	for i := 0; i < len(addresses); i++ {
		leaves[i] = keccak256(addresses[i], scores[i])
		if addresses[i] == targetAddress {
			targetAddress = leaves[i]
		}
	}

	sort.Strings(leaves)

	var proof []string
	for len(leaves) > 1 {
		var level []string
		for i := 0; i < len(leaves); i += 2 {
			left := leaves[i]
			var right string
			if i+1 < len(leaves) {
				right = leaves[i+1]
			} else {
				right = leaves[i]
			}

			if targetAddress == left {
				proof = append(proof, right)
				targetAddress = keccak256(left, right)
			}
			if targetAddress == right {
				proof = append(proof, left)
				targetAddress = keccak256(left, right)
			}

			level = append(level, keccak256(left, right))
		}
		leaves = level
	}

	return leaves[0], proof
}

func verify(proof []string, root, leaf string) bool {
	computedHash := leaf

	for _, proofElement := range proof {
		if computedHash <= proofElement {
			computedHash = keccak256(computedHash, proofElement)
		} else {
			computedHash = keccak256(proofElement, computedHash)
		}
	}

	return computedHash == root
}

func TestCreateProof(t *testing.T) {
	addresses := []string{
		"0x9D16512DD5b6C96E9E2196d30ff44F31Ca2d6077",
		"0x3770219B0F2ED1986E46FaE53b5D1A70d5a32eAb",
		"0xfF7d59D9316EBA168837E3eF924BCDFd64b237D8",
		"0x35405E1349658BcA12810d0f879Bf6c5d89B512C",
	}
	scores := []string{
		"200",
		"800",
		"100",
		"1000",
	}
	targetAddress := "0x3770219B0F2ED1986E46FaE53b5D1A70d5a32eAb"
	targetScore := "800"

	root, proof := buildMerkleTreeAndProof(addresses, scores, targetAddress)
	leaf := keccak256(targetAddress, targetScore)
	isValid := verify(proof, root, leaf)
	fmt.Println("Verification result:", isValid)

	fmt.Println("Merkle Root:", root)
	fmt.Println("Proof:[", strings.Join(proof, ","), "]")
	fmt.Println("Address: 0x", targetAddress)
	fmt.Println("Score:", targetScore)
	fmt.Println("Leaf:", leaf)
}
