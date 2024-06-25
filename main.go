package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "sync"
    "time"
)

type Client struct {
    conn net.Conn
    name string
}

var (
    clients   = make(map[net.Conn]Client)
    broadcast = make(chan string)
    mutex     = &sync.Mutex{}
)

func handleConnection(conn net.Conn) {
    defer conn.Close()

    fmt.Fprintln(conn, "Enter your name:")
    reader := bufio.NewReader(conn)
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)

    client := Client{conn: conn, name: name}

    mutex.Lock()
    clients[conn] = client
    mutex.Unlock()

    fmt.Fprintf(conn, "Welcome to the chat, %s!\n", name)
    broadcast <- fmt.Sprintf("%s has joined the chat!", name)

    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            
            delete(clients, conn)
            mutex.Unlock()
            broadcast <- fmt.Sprintf("%s has left the chat!", name)
            return
        }
        handleMessage(client, strings.TrimSpace(message))
    }
}

func handleMessage(client Client, message string) {
    if strings.HasPrefix(message, "/time") {
        currentTime := time.Now().Format("15:04:05")
        fmt.Fprintf(client.conn, "Current time: %s\n", currentTime)
    } else {
        broadcast <- fmt.Sprintf("%s: %s", client.name, message)
    }
}

func broadcastMessages() {
    for {
        msg := <-broadcast
        mutex.Lock()
        for _, client := range clients {
            fmt.Fprintln(client.conn, msg)
        }
        mutex.Unlock()
    }
}

func main() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer ln.Close()

    go broadcastMessages()

    fmt.Println("Chat server started on :8080")
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleConnection(conn)
    }
}
