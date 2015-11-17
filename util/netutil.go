package util

import (
	"net"	
	"fmt"
	"strconv"
	"encoding/binary"
	"github.com/lbarman/crypto/abstract"
	"encoding/hex"
)


func BroadcastMessage(conns []net.Conn, message []byte) {
	fmt.Println("Gonna broadcast this message")
	fmt.Println(hex.Dump(message))

	for i:=0; i<len(conns); i++ {
		n, err := conns[i].Write(message)

		fmt.Println("[", conns[i].LocalAddr(), " - ", conns[i].RemoteAddr(), "]")

		if n < len(message) || err != nil {
			fmt.Println("Could not broadcast to conn", i)
			panic("Error writing to socket:" + err.Error())
		}
	}
}

func TellPublicKey(conn net.Conn, LLD_PROTOCOL_VERSION int, publicKey abstract.Point) {
	publicKeyBytes, _ := publicKey.MarshalBinary()
	keySize := len(publicKeyBytes)

	//tell the relay our public key (assume user verify through second channel)
	buffer := make([]byte, 8+keySize)
	copy(buffer[8:], publicKeyBytes)
	binary.BigEndian.PutUint32(buffer[0:4], uint32(LLD_PROTOCOL_VERSION))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(keySize))

	n, err := conn.Write(buffer)

	if n < len(buffer) || err != nil {
		panic("Error writing to socket:" + err.Error())
	}
}
func MarshalPublicKeyArrayToByteArray(publicKeys []abstract.Point) []byte {
	var byteArray []byte

	msgType := make([]byte, 4)
	binary.BigEndian.PutUint32(msgType, uint32(2))
	byteArray = append(byteArray, msgType...)

	for i:=0; i<len(publicKeys); i++ {
		publicKeysBytes, err := publicKeys[i].MarshalBinary()
		publicKeyLength := make([]byte, 4)
		binary.BigEndian.PutUint32(publicKeyLength, uint32(len(publicKeysBytes)))

		byteArray = append(byteArray, publicKeyLength...)
		byteArray = append(byteArray, publicKeysBytes...)

		//fmt.Println(hex.Dump(publicKeysBytes))
		if err != nil{
			panic("can't marshal client public key n°"+strconv.Itoa(i))
		}
	}

	return byteArray
}

func UnMarshalPublicKeyArrayFromConnection(conn net.Conn, cryptoSuite abstract.Suite) []abstract.Point {

	println("Gonna read...")
	fmt.Println("[", conn.LocalAddr(), " - ", conn.RemoteAddr(), "]")

	fmt.Println(conn)
	//collect the public keys from the trustees
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		panic("Read error:" + err.Error())
	}
	println("OK")

	pks := UnMarshalPublicKeyArrayFromByteArray(buffer, cryptoSuite)
	return pks
}


func UnMarshalPublicKeyArrayFromByteArray(buffer []byte, cryptoSuite abstract.Suite) []abstract.Point {

	//will hold the public keys
	var publicKeys []abstract.Point

	//safety check
	messageType := int(binary.BigEndian.Uint32(buffer[0:4]))
	if messageType != 2 {
		panic("Trying to unmarshall an array, but does not start by 2")
	}

	//parse message
	currentByte := 4
	currentPkId := 0
	for {
		if currentByte+4 > len(buffer) {
			break; //we reached the end of the array
		}

		keyLength := int(binary.BigEndian.Uint32(buffer[currentByte:currentByte+4]))

		if keyLength == 0 {
			break; //we reached the end of the array
		}

		keyBytes := buffer[currentByte+4:currentByte+4+keyLength]

		publicKey := cryptoSuite.Point()
		err2 := publicKey.UnmarshalBinary(keyBytes)
		if err2 != nil {
			panic(">>>>can't unmarshal key n°"+strconv.Itoa(currentPkId)+" ! " + err2.Error())
		}

		publicKeys = append(publicKeys, publicKey)

		currentByte += 4 + keyLength
		currentPkId += 1
	}

	return publicKeys
}