package coco

import (
	"crypto/cipher"

	"github.com/dedis/crypto/abstract"
)

type SigningNode struct {
	Host
	suite   abstract.Suite
	pubKey  abstract.Point  // long lasting public key
	privKey abstract.Secret // long lasting private key

	v abstract.Secret // round lasting secret
	V abstract.Point  // round lasting commitment point
	c abstract.Secret // round lasting challenge
	r abstract.Secret // round lasting response

	V_hat abstract.Point  // aggregate of commit points
	r_hat abstract.Secret // aggregate of responses
	X_hat abstract.Point  // aggregate of public keys

	logTest  []byte                    // for testing purposes
	peerKeys map[string]abstract.Point // map of all peer public keys
}

func NewSigningNode(hn *HostNode, suite abstract.Suite, random cipher.Stream) *SigningNode {
	sn := &SigningNode{Host: hn, suite: suite}
	sn.privKey = suite.Secret().Pick(random)
	sn.pubKey = suite.Point().Mul(nil, sn.privKey)
	sn.X_hat = suite.Point().Null()
	return sn
}

func (sn *SigningNode) addPeer(conn Conn, pubKey abstract.Point) {
	sn.Host.AddPeers(conn)
	sn.peerKeys[conn.Name()] = pubKey
}
