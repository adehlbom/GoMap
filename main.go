package main

import (
	"fmt"
	"net"
	"os"

	"github.com/thediveo/netdb"
	"golang.org/x/exp/slices"
)

// vars
var helpvar int

func main() {
	init_script()
	checkFlags()

}
func init_script() {

}
func tcp_small() {
	var successful_ports []int
	arr := [5]int{20, 22, 80, 243, 3306}
	for i := 0; i < len(arr); i++ {
		address := fmt.Sprintf("192.168.0.104:%d", arr[i])
		conn, err := net.Dial("tcp", address)
		if err != nil {
			//fmt.Printf("Port %d is closed\n", arr[i])
			continue
		} else {
			successful_ports = append(successful_ports, arr[i])

		}
		conn.Close()
		//fmt.Printf("Port %d open\n", arr[i])
		what_service := netdb.ServiceByPort(arr[i], "tcp")
		fmt.Println(fmt.Sprint(what_service.Port) + " " + what_service.Name + " OPEN")
	}

}
func tcp_full() {
	var successful_ports []int
	for i := 1; i < 65535; i++ {
		address := fmt.Sprintf("192.168.0.104:%d", i)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			//fmt.Printf("Port %d is closed\n", arr[i])
			continue
		} else {
			successful_ports = append(successful_ports, i)

		}
		conn.Close()
		//fmt.Printf("Port %d open\n", arr[i])
		what_service := netdb.ServiceByPort(i, "tcp")
		fmt.Println(fmt.Sprint(what_service.Port) + " " + what_service.Name + " OPEN")

	}
}
func checkFlags() {
	if slices.Contains(os.Args, "-s") {
		tcp_small()
	} else if slices.Contains(os.Args, "-f") {
		tcp_full()
	}
}
