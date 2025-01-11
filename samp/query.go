package samp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

func Send(host, password, command string) (string, error) {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		return "", err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	request := new(bytes.Buffer)

	binary.Write(request, binary.LittleEndian, []byte("SAMP"))
	binary.Write(request, binary.LittleEndian, addr.IP.To4())
	binary.Write(request, binary.LittleEndian, uint16(addr.Port))
	binary.Write(request, binary.LittleEndian, uint8('x'))

	if err := binary.Write(request, binary.LittleEndian, uint16(len(password))); err != nil {
		return "", err
	}
	if err := binary.Write(request, binary.LittleEndian, []byte(password)); err != nil {
		return "", err
	}

	if err := binary.Write(request, binary.LittleEndian, uint16(len(command))); err != nil {
		return "", err
	}
	if err := binary.Write(request, binary.LittleEndian, []byte(command)); err != nil {
		return "", err
	}

	_, err = conn.Write(request.Bytes())
	if err != nil {
		return "", err
	}

	response := make([]byte, 2048)
	n, err := conn.Read(response)
	if err != nil {
		return "", err
	}
	if n > cap(response) {
		return "", errors.New("read response over buffer capacity")
	}

	var size uint16
	body := bytes.NewBuffer(response)
	body.Next(11)
	binary.Read(body, binary.LittleEndian, &size)
	buffer := make([]byte, size)
	binary.Read(body, binary.LittleEndian, &buffer)

	return string(buffer), nil
}
