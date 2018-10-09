package main

import (
	"encoding/base64"
	"io"
	"fmt"
	"encoding/binary"
	"log"
	"net"
    "golang.org/x/crypto/nacl/secretbox"
    "crypto/rand"
)

func main() {
	ourNonce := make([]byte, 24)
	rand.Read(ourNonce)
	fmt.Println("ourNonce", ourNonce)

	conn, err := net.Dial("tcp", "127.0.0.1:5199")
	if err != nil {
		log.Fatal(err)
	}
	n, err := conn.Write(ourNonce)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n, "bytes")
	theirNonce := make([]byte, 24)
	n2, err := io.ReadFull(conn, theirNonce)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got", n2, "bytes")
	fmt.Println("theirNonce", theirNonce)

	if len(ourNonce) != len(theirNonce) {
		log.Fatal("Received a nonce of size", len(theirNonce),",  expecting ", len(ourNonce))
	}

	var readingNonce [24]byte
	copy(readingNonce[0:12], ourNonce[0:12])
	copy(readingNonce[12:], theirNonce[12:])
	fmt.Println("readingNonce", readingNonce)

	var writingNonce [24]byte
	copy(writingNonce[0:12], theirNonce[0:12])
	copy(writingNonce[12:], ourNonce[12:])
	fmt.Println("writingNonce", writingNonce)

	command := []byte("print(123); return showVersion()")
	var key [32]byte
	xkey, err := base64.StdEncoding.DecodeString("WQcBTlKzEuTbMTdydMSW1CSQvyIAINML6oIGfGOjXjE=")
	if err != nil {
		log.Fatal(err)
	}
	copy(key[0:32], xkey)
	fmt.Println("key", key)
	encodedcommand := make([]byte, 0)
	encodedcommand = secretbox.Seal(encodedcommand, command, &writingNonce, &key)

	fmt.Println("encodedcommand", encodedcommand)
	sendlen := make([]byte, 4)
	binary.BigEndian.PutUint32(sendlen, uint32(len(encodedcommand)))
	n3, err := conn.Write(sendlen)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n3, "bytes")
	n4, err := conn.Write(encodedcommand)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("wrote", n4, "bytes")

	recvlenbuf := make([]byte, 4)
	n5, err := io.ReadFull(conn, recvlenbuf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("read", n5, "bytes")
	recvlen := binary.BigEndian.Uint32(recvlenbuf)
	fmt.Println("should read", recvlen, "bytes")
	recvbuf := make([]byte, recvlen)
	n6, err := io.ReadFull(conn, recvbuf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("read", n6, "bytes")
	decodedresponse := make([]byte, 0)
	decodedresponse, ok := secretbox.Open(decodedresponse, recvbuf, &readingNonce, &key)
	if !ok {
		log.Fatal("secretbox")
	}
	fmt.Println("response:", string(decodedresponse))
}