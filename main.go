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
	start := time.Now()

	fmt.Println("Starting GoMap (https://github.com/adehlbom/GoMap) at " + time.Now().Format("2006-01-02 15:04:05"))
	ip_address := "192.168.0.100"
	end_port := checkFlags()
	for port := 1; port <= end_port; port++ {
		wg.Add(1)
		go tcp_scan(ip_address, port)

	}
	wg.Wait()
	fmt.Printf("All scanned. This took %ds\n", time.Since(start)/1000000000)

}
func tcp_scan(ip_address string, port int) {
	defer wg.Done()
	address := fmt.Sprintf("%s:%d", ip_address, port)
	conn, err := net.DialTimeout("tcp", address, 15*time.Second)
	if err != nil {
		if port == 22 || port == 80 || port == 53 {
			fmt.Println(err)

		}
		//conn.Close()
		return

	}
	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	numbytesread, err2 := conn.Read(buffer)

	if err2 != nil {
		fmt.Println(err2)

		//conn.Close()
		return
	}
	fmt.Printf("%d/tcp OPEN | SERVICE: %s \n ", port, string(buffer[0:numbytesread]))
	//fmt.Println(string(buffer[0:numbytesread]))
	conn.Close()
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
