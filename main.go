package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
)

var wg sync.WaitGroup

func main() {
	fmt.Println("Starting GoMap (https://github.com/adehlbom/GoMap) at " + time.Now().Local().String())
	ip_address := "45.33.32.156"
	end_port := checkFlags()
	for port := 1; port <= end_port; port++ {
		wg.Add(1)
		go tcp_scan(ip_address, port)

	}
	wg.Wait()

	fmt.Println("All scanned.")

}
func tcp_scan(ip_address string, port int) string {
	defer wg.Done()
	address := fmt.Sprintf("%s:%d", ip_address, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		fmt.Println(err)
		//conn.Close()
		return ""
	}
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	numbytesread, err2 := conn.Read(buffer)
	if err2 != nil {
		fmt.Println(err2)
		//conn.Close()
		return ""
	}
	fmt.Printf("Port %d open\n", port)
	fmt.Println(string(buffer[0:numbytesread]))
	conn.Close()

	return string(buffer[0:numbytesread])
	// Hur och var använder jag defer conn.Close här när jag har flera if-satser som körs efter varandra?
	//fmt.Println(string(buffer[0:numbytesread]))

	//what_service := netdb.ServiceByPort(i, "tcp")
	//fmt.Println(fmt.Sprint(what_service.Port) + " " + what_service.Name + " OPEN")

}
func test_ssh(ip_address string, port int) {
	fmt.Println("Testing to connect to SSH server...")
	var hostKey ssh.PublicKey

	config := &ssh.ClientConfig{
		User: "",
		Auth: []ssh.AuthMethod{
			ssh.Password(""),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip_address, port), config)

	if err != nil {
		fmt.Println(err)
	}
	conn.Close()

}

func checkFlags() int {
	if slices.Contains(os.Args, "-s") {
		return 1024
	} else if slices.Contains(os.Args, "-f") {
		return 65535
	} else if slices.Contains(os.Args, "-ip") {

	} else if slices.Contains(os.Args, "-h") {

	}
	return 0
}
