package cmd

import (
	"fmt"
	"github.com/leancodebox/goose/array"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"net"
	"strings"
	"sync"
	"time"
)

func init() {
	cmd := &cobra.Command{
		Use:   "tcp_scan",
		Short: "tcp_scan --ip=127.0.0.1 --port1=80 --port2=81",
		Run:   runTcpScan,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().String("ip", "127.0.0.1", "IP address")
	cmd.Flags().Int("port1", 1, "")
	cmd.Flags().Int("port2", 65535, "")
	appendCommand(cmd)
}
func runTcpScan(cmd *cobra.Command, _ []string) {
	target, _ := cmd.Flags().GetString("ip") // 目标主机地址
	port1, _ := cmd.Flags().GetInt("port1")  // 目标主机地址
	port2, _ := cmd.Flags().GetInt("port2")  // 目标主机地址
	timeout := 5                             // 超时时间（秒）

	var wg sync.WaitGroup
	maxConcurrency := 100 // 最大并发数限制，可根据需要调整
	// 创建信号量，用于控制并发数量
	semaphore := make(chan struct{}, maxConcurrency)
	for port := port1; port <= port2; port++ {
		wg.Add(1)
		semaphore <- struct{}{} // 获取一个信号量，代表一个 Goroutine 占用一个并发槽位
		go func(ip, port any) {
			defer func() {
				<-semaphore // 释放信号量，代表一个 Goroutine 释放一个并发槽位
				wg.Done()
			}()
			addr := fmt.Sprintf("%s:%d", target, port)
			conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
			if err != nil { // 端口不可访问
				return
			}

			defer conn.Close()
			fmt.Printf("%s:%d is open\n", ip, port)
		}(target, port)
	}

	wg.Wait()
	fmt.Println("Scan completed")
}

func init() {
	cmd := &cobra.Command{
		Use:   "tcp_scan_list",
		Short: "tcp_scan_list --ip1=10.249.1.1 --ip2=10.249.255.255 --port=80,81",
		Run:   runTcpScan2,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().String("ip1", "127.0.0.1", "IP address 1")
	cmd.Flags().String("ip2", "127.0.0.1", "IP address 2")
	cmd.Flags().String("port", "22,80", "port number 22,80")
	appendCommand(cmd)
}

func runTcpScan2(cmd *cobra.Command, _ []string) {
	start := time.Now()
	startIP, _ := cmd.Flags().GetString("ip1")
	endIP, _ := cmd.Flags().GetString("ip2")
	portListStr, _ := cmd.Flags().GetString("port")

	//startIP = "10.249.1.1"
	//endIP = "10.249.255.255"
	port := array.ArrayMap(func(t string) int {
		return cast.ToInt(t)
	}, strings.Split(portListStr, ","))

	//port := []int{ /*22, 80,*/ 8081, 8080}
	scanIPRange(startIP, endIP, port)

	elapsed := time.Since(start)
	fmt.Printf("\nScan completed in %v\n", elapsed)
}

func scanIPRange(startIP, endIP string, ports []int) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start.To4() == nil || end.To4() == nil {
		fmt.Println("Invalid IP address")
		return
	}

	startInt := ipToInt(start.To4())
	endInt := ipToInt(end.To4())

	if startInt > endInt {
		fmt.Println("Invalid IP range")
		return
	}

	var wg sync.WaitGroup
	maxConcurrency := 100 // 最大并发数限制，可根据需要调整

	// 创建信号量，用于控制并发数量
	semaphore := make(chan struct{}, maxConcurrency)

	// 遍历 IP 段中的所有 IP 地址，依次进行端口扫描
	for i := startInt; i <= endInt; i++ {
		ip := intToIP(i)

		wg.Add(len(ports))

		// 对指定 IP 地址上的所有端口依次进行扫描
		for _, port := range ports {
			semaphore <- struct{}{} // 获取一个信号量，代表一个 Goroutine 占用一个并发槽位
			go func(ip string, port int) {
				defer func() {
					<-semaphore // 释放信号量，代表一个 Goroutine 释放一个并发槽位
					wg.Done()
				}()
				target := fmt.Sprintf("%s:%d", ip, port)
				conn, err := net.DialTimeout("tcp", target, 200*time.Millisecond)
				if err != nil { // 端口不可访问
					return
				}
				defer conn.Close()
				fmt.Printf("%s:%d is open\n", ip, port)
			}(ip, port)
		}
	}

	wg.Wait()

	fmt.Println("\nScan completed")
}

func ipToInt(ip net.IP) int64 {
	if len(ip) == 16 {
		return int64(ip[12])<<24 | int64(ip[13])<<16 | int64(ip[14])<<8 | int64(ip[15])
	}
	return int64(ip[0])<<24 | int64(ip[1])<<16 | int64(ip[2])<<8 | int64(ip[3])
}

func intToIP(i int64) string {
	ip := make(net.IP, 4)
	ip[0] = byte(i >> 24 & 0xFF)
	ip[1] = byte(i >> 16 & 0xFF)
	ip[2] = byte(i >> 8 & 0xFF)
	ip[3] = byte(i & 0xFF)
	return ip.String()
}
