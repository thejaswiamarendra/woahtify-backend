package api

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"woahtify-backend/utils"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	roomConfigMap    map[string]*RoomConfig
	roomBroadcastMap map[string]chan []byte
	mutex            *sync.Mutex
}

type WSUser struct {
	UserName string `json:"userName"`
	UserType string `json:"userType"`
	IsAlive  bool   `json:"isAlive"`
}

func (u *WSUser) isEqual(user WSUser) bool {
	return u.UserName == user.UserName && u.UserType == user.UserType && u.IsAlive == user.IsAlive
}

func removeUserFromList(users []*WSUser, target WSUser) []*WSUser {
	result := []*WSUser{}
	for _, user := range users {
		if user != &target {
			result = append(result, user)
		}
	}
	return result
}

type BroadcastMessage struct {
	RoomName          string        `json:"roomname"`
	Sender            WSUser        `json:"sender"`
	Message           string        `json:"message"`
	CurrentSongQueue  []*SongConfig `json:"currentSongQueue"`
	CurrentSong       *SongConfig   `json:"currentSong"`
	ConnectedUserList []*WSUser     `json:"connectedUserList"`
	ConnectionID      string        `json:"connectionID"`
}

type SongConfig struct {
	SongName           string    `json:"songName"`
	Votes              []*WSUser `json:"votes"`
	VoteCount          int       `json:"voteCount"`
	SuggestedBy        WSUser    `json:"suggestedBy"`
	SuggestedTimestamp time.Time `json:"suggestedTimeStamp"`
	Index              int       `json:"index"`
}

type SongPriorityQueue []*SongConfig

func (sp SongPriorityQueue) Len() int {
	return len(sp)
}

func (sp SongPriorityQueue) Less(i, j int) bool {
	if sp[i].VoteCount != sp[j].VoteCount {
		// Higher the VoteCount, higher the priority
		return sp[i].VoteCount > sp[j].VoteCount
	}
	if !sp[i].SuggestedTimestamp.Equal(sp[j].SuggestedTimestamp) {
		// Lower the timestamp, higher the second priority
		return sp[i].SuggestedTimestamp.Before(sp[j].SuggestedTimestamp)
	}
	// Lower the ASCII, higher the third priority
	return sp[i].SongName < sp[j].SongName
}

func (sp *SongPriorityQueue) Pop() interface{} {
	oldSpq := *sp
	n := len(oldSpq)
	song := oldSpq[n-1]
	song.Index = -1
	*sp = oldSpq[0 : n-1]
	return song
}

func (sp *SongPriorityQueue) Push(songToPush interface{}) {
	n := len(*sp)
	song := songToPush.(*SongConfig)
	song.Index = n
	*sp = append(*sp, song)
}

func (sp SongPriorityQueue) Swap(i, j int) {
	sp[i], sp[j] = sp[j], sp[i]
	sp[i].Index = i
	sp[j].Index = j
}

func (sp *SongPriorityQueue) update(currentSong *SongConfig, votes []*WSUser) {
	currentSong.Votes = votes
	currentSong.VoteCount = len(votes)
	heap.Fix(sp, currentSong.Index)
}

type RoomConfig struct {
	Host                WSUser
	IsHostPresent       bool
	RoomName            string
	Clients             map[*websocket.Conn]WSUser
	ConnectionIDUserMap map[string]*websocket.Conn
	ConnectedUserList   []*WSUser
	SongQueue           SongPriorityQueue
	CurrentSong         *SongConfig
	Secret              string
}

func NewWSServer() *WSServer {
	return &WSServer{
		roomConfigMap:    make(map[string]*RoomConfig),
		roomBroadcastMap: make(map[string]chan []byte),
		mutex:            &sync.Mutex{},
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws *WSServer) isRoomPresent(roomName string) bool {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	_, exists := ws.roomConfigMap[roomName]
	return exists
}

func (ws *WSServer) addRoom(roomName string, host WSUser) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if _, exists := ws.roomConfigMap[roomName]; exists {
		return fmt.Errorf("room %s already present", roomName)
	}
	secret, err := utils.GenerateSecureRandomString(32)
	if err != nil {
		return err
	}

	ws.roomConfigMap[roomName] = &RoomConfig{
		Host:                host,
		IsHostPresent:       false,
		RoomName:            roomName,
		Clients:             make(map[*websocket.Conn]WSUser),
		ConnectionIDUserMap: make(map[string]*websocket.Conn),
		SongQueue:           SongPriorityQueue{},
		CurrentSong:         nil,
		ConnectedUserList:   []*WSUser{},
		Secret:              secret,
	}
	heap.Init(&ws.roomConfigMap[roomName].SongQueue)
	ws.roomBroadcastMap[roomName] = make(chan []byte, 16)
	go ws.roomBroadcaster(roomName)
	return nil
}

// joinUser atomically checks conditions and adds a user to a room.
// It prevents race conditions by performing all checks and modifications within a single lock.
func (ws *WSServer) joinUser(roomName string, userName string, conn *websocket.Conn) (WSUser, string, error) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		return WSUser{}, "", fmt.Errorf("room '%s' not found", roomName)
	}

	// Check for duplicate username
	for _, existingUser := range room.Clients {
		if existingUser.UserName == userName {
			return WSUser{}, "", fmt.Errorf("user '%s' is already present in the room", userName)
		}
	}

	var userType string
	if userName == room.Host.UserName {
		if room.IsHostPresent {
			return WSUser{}, "", fmt.Errorf("host already present in room '%s'", roomName)
		}
		userType = "host"
	} else {
		if !room.IsHostPresent {
			return WSUser{}, "", fmt.Errorf("host is not yet present in room '%s', please wait", roomName)
		}
		userType = "guest"
	}

	connID, err := utils.GenerateSecureRandomString(32)
	if err != nil {
		return WSUser{}, "", err
	}
	encryptedConnID, err := utils.Encrypt(connID, room.Secret)

	if err != nil {
		return WSUser{}, "", err
	}

	user := WSUser{
		UserName: userName,
		UserType: userType,
		IsAlive:  true,
	}

	if user.UserType == "host" {
		room.IsHostPresent = true
	}

	room.Clients[conn] = user
	room.ConnectionIDUserMap[connID] = conn
	room.ConnectedUserList = append(room.ConnectedUserList, &user)

	ws.broadcastUpdate(roomName, encryptedConnID)
	return user, encryptedConnID, nil
}

func (ws *WSServer) removeUser(roomName, connectionID string, conn *websocket.Conn) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		log.Printf("Room %s not present", roomName)
		return fmt.Errorf("room %s not present", roomName)
	}

	user, userExists := room.Clients[conn]
	if !userExists {
		log.Printf("User not present")
		return fmt.Errorf("user not present")
	}

	decryptedConnID, err := utils.Decrypt(connectionID, room.Secret)
	if err != nil {
		// Log the error but proceed to remove the user by connection object if possible
		log.Printf("Could not decrypt connectionID for user %s: %v", user.UserName, err)
	}

	delete(room.Clients, conn)
	delete(room.ConnectionIDUserMap, decryptedConnID)
	room.ConnectedUserList = removeUserFromList(room.ConnectedUserList, user)
	log.Printf("User %s removed from room %s\n", user.UserName, roomName)

	if user.UserType == "host" {
		log.Printf("Host left room %s. Deleting room.\n", roomName)
		close(ws.roomBroadcastMap[roomName])
		delete(ws.roomBroadcastMap, roomName)
		delete(ws.roomConfigMap, roomName)
	} else {
		ws.broadcastUpdate(roomName, connectionID)
	}
	return nil
}

func (ws *WSServer) addSuggestedSong(songName, roomName, connectionID string) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		log.Printf("Room %s not present", roomName)
		return fmt.Errorf("room %s not present", roomName)
	}

	for _, song := range room.SongQueue {
		if song.SongName == songName {
			log.Printf("Song already suggested by %s", song.SuggestedBy.UserName)
			return fmt.Errorf("song already suggested by %s", song.SuggestedBy.UserName)
		}
	}

	decryptedConnId, err := utils.Decrypt(connectionID, room.Secret)

	if err != nil {
		return err
	}

	conn, exists := room.ConnectionIDUserMap[decryptedConnId]
	if !exists {
		log.Printf("Connection with connectionID %s doesn't exist", connectionID)
		return fmt.Errorf("connection with connection id %s doesn't exist", connectionID)
	}
	user := room.Clients[conn]

	if len(room.SongQueue) == 0 && room.CurrentSong == nil {
		room.CurrentSong = &SongConfig{
			SongName:           songName,
			Votes:              []*WSUser{&user},
			VoteCount:          1,
			SuggestedBy:        user,
			SuggestedTimestamp: time.Now(),
		}
		ws.broadcastUpdate(roomName, connectionID)
		return nil
	}

	heap.Push(
		&room.SongQueue,
		&SongConfig{
			SongName:           songName,
			Votes:              []*WSUser{&user},
			VoteCount:          1,
			SuggestedBy:        user,
			SuggestedTimestamp: time.Now(),
		},
	)

	ws.broadcastUpdate(roomName, connectionID)
	return nil
}

func (ws *WSServer) voteForSong(songName, roomName, connectionID string) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		log.Printf("Room %s not present", roomName)
		return fmt.Errorf("room %s not present", roomName)
	}

	decryptedConnId, err := utils.Decrypt(connectionID, room.Secret)

	if err != nil {
		return err
	}

	conn, exists := room.ConnectionIDUserMap[decryptedConnId]
	if !exists {
		log.Printf("Connection with connectionID %s doesn't exist", connectionID)
		return fmt.Errorf("connection with connection id %s doesn't exist", connectionID)
	}
	user := room.Clients[conn]

	for _, song := range room.SongQueue {
		if song.SongName == songName {
			for _, u := range song.Votes {
				if u.isEqual(user) {
					log.Printf("User %s has already voted for song %s", user.UserName, songName)
					return fmt.Errorf("user %s has already voted for song %s", user.UserName, songName)
				}
			}
			votes := append(song.Votes, &user)
			ws.roomConfigMap[roomName].SongQueue.update(song, votes)
			ws.broadcastUpdate(roomName, connectionID)
			log.Printf("User %s cast a vote for the song %s", user.UserName, songName)
			return nil
		}
	}

	log.Printf("Song hasn't been suggested")
	return fmt.Errorf("song hasn't been suggested")
}

func (ws *WSServer) skipSong(songName, roomName, connectionID string) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		log.Printf("Room %s not present", roomName)
		return fmt.Errorf("room %s not present", roomName)
	}

	decryptedConnId, err := utils.Decrypt(connectionID, room.Secret)

	if err != nil {
		return err
	}

	conn, exists := room.ConnectionIDUserMap[decryptedConnId]
	if !exists {
		log.Printf("Connection with connectionID %s doesn't exist", connectionID)
		return fmt.Errorf("connection with connection id %s doesn't exist", connectionID)
	}
	user := room.Clients[conn]
	if user.UserType != "host" {
		log.Printf("Only a host can skip a song, %s", user.UserType)
		return fmt.Errorf("only a host can skip a song")
	}
	if songName != room.CurrentSong.SongName {
		log.Printf("Can't skip a song that is not playing")
		return fmt.Errorf("can't skip a song that is not playing")
	}
	room.CurrentSong = nil
	if len(room.SongQueue) == 0 {
		log.Printf("no more songs in the queue")
		return fmt.Errorf("no more songs in the queue")
	}

	nextSong := heap.Pop(&room.SongQueue).(*SongConfig)
	room.CurrentSong = nextSong
	log.Printf("Skipped %s, playing %s", songName, nextSong.SongName)
	ws.broadcastUpdate(roomName, connectionID)
	return nil
}

func (ws *WSServer) handleClientMessages(roomName, connectionID string, conn *websocket.Conn) {
	defer func() {
		ws.removeUser(roomName, connectionID, conn)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client in room %s: %v\n", roomName, err)
			break
		}

		ws.mutex.Lock()
		room, roomExists := ws.roomConfigMap[roomName]
		if !roomExists {
			ws.mutex.Unlock()
			break
		}
		broadCaseMessage := BroadcastMessage{
			Message:           string(message),
			RoomName:          roomName,
			Sender:            room.Clients[conn],
			CurrentSongQueue:  room.SongQueue,
			ConnectedUserList: room.ConnectedUserList,
		}
		marshalledMessage, err := json.Marshal(broadCaseMessage)
		if err != nil {
			log.Printf("Error marshaling message: %v\n", err)
			ws.mutex.Unlock()
			continue
		}

		if broadcastChan, ok := ws.roomBroadcastMap[roomName]; ok {
			select {
			case broadcastChan <- marshalledMessage:
			default:
				log.Printf("Warning: broadcast channel for room %s is full. Message from user dropped.", roomName)
			}
		}
		ws.mutex.Unlock()
	}
}

// broadcastUpdate must be called with the mutex held.
// It constructs the current state and sends it to the room's broadcast channel.
func (ws *WSServer) broadcastUpdate(roomName, connectionID string) {
	room, roomExists := ws.roomConfigMap[roomName]
	if !roomExists {
		return
	}
	log.Print("In BroadCast Message")
	for _, song := range room.SongQueue {
		log.Printf("Song %s, Votes %v", song.SongName, song.VoteCount)
	}
	conn := room.ConnectionIDUserMap[connectionID]
	sender := room.Clients[conn]
	log.Printf("Connection %s User %s for connection ID %+v", connectionID, sender.UserName, conn)
	broadcastMessage := BroadcastMessage{
		Sender:            sender,
		RoomName:          roomName,
		CurrentSongQueue:  room.SongQueue,
		CurrentSong:       room.CurrentSong,
		ConnectedUserList: room.ConnectedUserList,
		ConnectionID:      connectionID,
	}

	marshalledMessage, err := json.Marshal(broadcastMessage)
	if err != nil {
		log.Printf("Error marshaling update message: %v\n", err)
		return
	}

	// Non-blocking send to avoid deadlocking if the broadcast channel is full.
	// This is critical because this function is called while holding the server-wide mutex.
	// A blocking send here would halt all other operations on the WSServer.
	select {
	case ws.roomBroadcastMap[roomName] <- marshalledMessage:
	default:
		log.Printf("Warning: broadcast channel for room %s is full. Update dropped.", roomName)
	}
}

func (ws *WSServer) roomBroadcaster(roomName string) {
	broadcastChan, ok := ws.roomBroadcastMap[roomName]
	if !ok {
		return
	}

	for msg := range broadcastChan {
		ws.mutex.Lock()
		room, ok := ws.roomConfigMap[roomName]
		if !ok {
			ws.mutex.Unlock()
			return // Room was closed.
		}

		// Copy client connections to a slice to avoid holding the lock during I/O.
		clients := make([]*websocket.Conn, 0, len(room.Clients))
		for c := range room.Clients {
			clients = append(clients, c)
		}
		ws.mutex.Unlock()

		for _, c := range clients {
			c.WriteMessage(websocket.TextMessage, msg)
		}
	}
}
