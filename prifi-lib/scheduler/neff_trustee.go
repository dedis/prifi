package scheduler

/*

The interface should be :

INPUT : list of client's public keys

OUTPUT : list of slots

*/

import (
	"bytes"
	"errors"
	"github.com/dedis/crypto/abstract"
	crypto_proof "github.com/dedis/crypto/proof"
	"github.com/dedis/crypto/random"
	"github.com/dedis/crypto/shuffle"
	"github.com/lbarman/prifi/prifi-lib"
	"github.com/lbarman/prifi/prifi-lib/config"
	"github.com/lbarman/prifi/prifi-lib/crypto"
	"math/rand"
	"strconv"
)

type neffShuffleTrusteeView struct {
	TrusteeId  int
	PrivateKey abstract.Scalar
	PublicKey  abstract.Point

	SecretCoeff   abstract.Scalar // c[i]
	Shares        abstract.Scalar // s[i] = c[0] * ... c[1]
	Proof         []byte
	EphemeralKeys []abstract.Point
}

/**
 * Creates a new trustee-view for the neff shuffle, and initiates the fields correctly
 */
func (t *neffShuffleTrusteeView) init(trusteeId int, private abstract.Scalar, public abstract.Point) error {
	if trusteeId < 0 {
		return errors.New("Cannot shuffle without a valid id (>= 0)")
	}
	if private == nil {
		return errors.New("Cannot shuffle without a private key.")
	}
	if public == nil {
		return errors.New("Cannot shuffle without a public key.")
	}
	t.TrusteeId = trusteeId
	t.PrivateKey = private
	t.PublicKey = public
	return nil
}

/**
 * Received s[i-1], and the public keys. Do the shuffle, store locally, and send back the new s[i], shuffle array
 * If shuffleKeyPositions is false, do not shuffle the key's position (useful for testing - 0 anonymity)
 */
func (t *neffShuffleTrusteeView) ReceivedShuffleFromRelay(lastShares abstract.Scalar, clientPublicKeys []abstract.Point, shuffleKeyPositions bool) (error, interface{}) {

	if lastShares == nil {
		return errors.New("Cannot perform a shuffle is lastShare is nil"), nil
	}
	if clientPublicKeys == nil {
		return errors.New("Cannot perform a shuffle is clientPublicKeys is nil"), nil
	}
	if len(clientPublicKeys) == 0 {
		return errors.New("Cannot perform a shuffle is len(clientPublicKeys) is 0"), nil
	}

	//compute new shares
	secretCoeff := config.CryptoSuite.Scalar().Pick(random.Stream)
	t.SecretCoeff = secretCoeff
	newShares := config.CryptoSuite.Scalar().Mul(lastShares, secretCoeff)

	//transform the public keys with the secret coeff
	ephPublicKeys2 := clientPublicKeys
	for i := 0; i < len(clientPublicKeys); i++ {
		oldKey := clientPublicKeys[i]
		ephPublicKeys2[i] = config.CryptoSuite.Point().Mul(oldKey, secretCoeff)
	}

	//shuffle the array
	if shuffleKeyPositions {
		//TODO : I'm not shure this actually shuffles ?
		ephPublicKeys3 := make([]abstract.Point, len(ephPublicKeys2))
		perm := rand.Perm(len(ephPublicKeys2))
		for i, v := range perm {
			ephPublicKeys3[v] = ephPublicKeys2[i]
		}
		ephPublicKeys2 = ephPublicKeys3
	}

	proof := make([]byte, 50) // TODO : the proof should be done

	//store the result
	t.Shares = newShares
	t.EphemeralKeys = ephPublicKeys2
	t.Proof = proof

	//send the answer
	msg := &prifi_lib.TRU_REL_TELL_NEW_BASE_AND_EPH_PKS{
		NewBase:   newShares,
		NewEphPks: ephPublicKeys2,
		Proof:     proof}

	return nil, msg
}

/**
 * We received a transcript of the whole shuffle from the relay. Check that we are included, and sign
 */
func (t *neffShuffleTrusteeView) ReceivedTranscriptFromRelay(shares []abstract.Scalar, shuffledPublicKeys [][]abstract.Point, proofs [][]byte) (error, interface{}) {

	if t.Shares == nil {
		return errors.New("Cannot verify the shuffle, we didn't store the base"), nil
	}
	if t.EphemeralKeys == nil || len(t.EphemeralKeys) == 0 {
		return errors.New("Cannot verify the shuffle, we didn't store the ephemeral keys"), nil
	}
	if t.Proof == nil {
		return errors.New("Cannot verify the shuffle, we didn't store the proof"), nil
	}
	if len(shares) != len(shuffledPublicKeys) || len(shares) != len(proofs) {
		return errors.New("Size not matching, G_s is " + strconv.Itoa(len(shares)) + ", shuffledPublicKeys_s is " + strconv.Itoa(len(shuffledPublicKeys)) + ", proof_s is " + strconv.Itoa(len(proofs)) + "."), nil
	}

	nTrustees := len(shares)
	nClients := len(shuffledPublicKeys[0])

	//Todo : verify each individual permutations. No verification is done yet
	var err error
	for j := 0; j < nTrustees; j++ {

		verify := true
		if j > 0 {
			X := shuffledPublicKeys[j-1]
			Y := shuffledPublicKeys[j-1]
			Xbar := shuffledPublicKeys[j]
			Ybar := shuffledPublicKeys[j]
			if len(X) > 1 {
				verifier := shuffle.Verifier(config.CryptoSuite, nil, X[0], X, Y, Xbar, Ybar)
				err = crypto_proof.HashVerify(config.CryptoSuite, "PairShuffle", verifier, proofs[j])
			}
			if err != nil {
				verify = false
			}
		}
		verify = true // TODO: This shuffle needs to be fixed

		if !verify {
			return errors.New("Could not verify the " + strconv.Itoa(j) + "th neff shuffle, error is " + err.Error()), nil
		}
	}

	//we verify that our shuffle was included
	ownPermutationFound := false
	for j := 0; j < nTrustees; j++ {
		if shares[j].Equal(t.Shares) && bytes.Equal(t.Proof, proofs[j]) {
			allKeyEqual := true
			for k := 0; k < nClients; k++ {
				if !t.EphemeralKeys[k].Equal(shuffledPublicKeys[j][k]) {
					allKeyEqual = false
					break
				}
			}
			if allKeyEqual {
				ownPermutationFound = true
			}
		}
	}

	if !ownPermutationFound {
		return errors.New("Could not locate our own permutation in the transcript..."), nil
	}

	//prepare the transcript signature. Since it is OK, we're gonna sign only the latest permutation
	var blob []byte
	lastPerm := nTrustees - 1

	lastSharesByte, err := shares[lastPerm].MarshalBinary()
	if err != nil {
		return errors.New("Can't marshall the last shares..."), nil
	}
	blob = append(blob, lastSharesByte...)

	for j := 0; j < nClients; j++ {
		pkBytes, err := shuffledPublicKeys[lastPerm][j].MarshalBinary()
		if err != nil {
			return errors.New("Can't marshall shuffled public key" + strconv.Itoa(j)), nil
		}
		blob = append(blob, pkBytes...)
	}

	//sign this blob
	signature := crypto.SchnorrSign(config.CryptoSuite, random.Stream, blob, t.PrivateKey)

	//send the answer
	msg := &prifi_lib.TRU_REL_SHUFFLE_SIG{
		TrusteeID: t.TrusteeId,
		Sig:       signature}

	return nil, msg
}
