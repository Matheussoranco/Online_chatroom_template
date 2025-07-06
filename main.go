package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

// Configuração do WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite todas as origens (apenas para desenvolvimento!)
	},
}

// Cliente representa um usuário conectado
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// Gerenciador de clientes
var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

func main() {
	// Rota para WebSocket
	http.HandleFunc("/ws", handleWebSocket)

	// Goroutine para enviar mensagens para todos os clientes
	go broadcastMessages()

	// Inicia o servidor
	log.Println("Servidor de chat iniciado na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleWebSocket lida com conexões WebSocket
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao atualizar para WebSocket:", err)
		return
	}
	defer conn.Close()

	// Cria um novo cliente
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	clients[client] = true

	// Loop para ler mensagens do cliente
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Erro ao ler mensagem:", err)
			delete(clients, client)
			break
		}
		broadcast <- message
	}
}

// broadcastMessages envia mensagens para todos os clientes
func broadcastMessages() {
	for {
		message := <-broadcast
		for client := range clients {
			err := client.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Erro ao enviar mensagem:", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}