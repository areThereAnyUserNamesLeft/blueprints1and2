package main

import (
	"log"
	"net/http"
	"playground/23052018/chat/trace"

	"github.com/gorilla/websocket"
)

type room struct {
	// forward is a channel of incoming messages
	forward chan []byte
	// join is a channel for clients wanting to join a room
	join chan *client
	// leave is a channel for clients wanting to leave a room
	leave chan *client
	// Clients holds all the clients current in the room
	clients map[*client]bool
	// tracer will recieve the trace info of activity in the room
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join: //joining
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave: // leaving
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("client left")
		case msg := <-r.forward: // forward to all
			for client := range r.clients {
				client.send <- msg // send message
				r.tracer.Trace(" -- sent to client")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
