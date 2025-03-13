// main.go - 一個簡單的 TCP 聊天伺服器
package main

import (
	"bufio"   // 用於讀取輸入
	"fmt"     // 用於格式化輸入輸出
	"net"     // 提供網路功能
	"strings" // 字串處理
	"sync"    // 提供同步原語，如互斥鎖
)

// Client 結構體表示一個連接到伺服器的客戶端
type Client struct {
	conn net.Conn // 客戶端的網路連接
	name string   // 客戶端的名稱
}

var (
	clients   = make(map[net.Conn]Client) // 儲存所有連接的客戶端
	broadcast = make(chan string)         // 廣播訊息的通道
	mutex     = &sync.Mutex{}             // 互斥鎖，用於保護 clients 映射的並發訪問
)

// handleConnection 處理單個客戶端連接
func handleConnection(conn net.Conn) {
	defer conn.Close() // 確保連接在函數結束時關閉

	// 提示用戶輸入名稱
	fmt.Fprintln(conn, "Enter your name:")
	reader := bufio.NewReader(conn)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name) // 移除名稱中的空白字符

	// 創建新的客戶端並添加到 clients 映射中
	client := Client{conn: conn, name: name}

	mutex.Lock() // 加鎖以保護 clients 映射的並發訪問
	clients[conn] = client
	mutex.Unlock()

	// 歡迎新客戶端並通知其他客戶端
	fmt.Fprintf(conn, "Welcome to the chat, %s!\n", name)
	broadcast <- fmt.Sprintf("%s has joined the chat!", name)

	// 持續讀取客戶端發送的訊息
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// 如果讀取出錯，表示客戶端已斷開連接
			mutex.Lock()
			delete(clients, conn) // 從 clients 映射中移除客戶端
			mutex.Unlock()
			broadcast <- fmt.Sprintf("%s has left the chat!", name) // 通知其他客戶端
			return
		}
		// 廣播客戶端發送的訊息
		broadcast <- fmt.Sprintf("%s: %s", name, strings.TrimSpace(message))
	}
}

// broadcastMessages 將訊息廣播給所有連接的客戶端
func broadcastMessages() {
	for {
		msg := <-broadcast // 從廣播通道接收訊息
		mutex.Lock()       // 加鎖以保護 clients 映射的並發訪問
		for _, client := range clients {
			fmt.Fprintln(client.conn, msg) // 將訊息發送給每個客戶端
		}
		mutex.Unlock()
	}
}

// main 函數是程式的入口點
func main() {
	// 在 8080 端口上啟動 TCP 伺服器
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close() // 確保伺服器在函數結束時關閉

	// 啟動廣播訊息的 goroutine
	go broadcastMessages()

	fmt.Println("Chat server started on :8080")
	// 持續接受新的客戶端連接
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		// 為每個新連接啟動一個 goroutine 處理
		go handleConnection(conn)
	}
}
