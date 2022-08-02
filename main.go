package main

import (
	"fmt"
	"net"
	"https://github.com/google/gopacket"
)

func main() {
	arr := [4]int{20, 22, 80, 243}
	for i := 0; i < len(arr); i++ {
		address := fmt.Sprintf("scanme.nmap.org:%d", arr[i])
		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Printf("Port %d is closed\n", arr[i])
			continue
		}
		conn.Close()
		fmt.Printf("Port %d open\n", arr[i])
	}

}
