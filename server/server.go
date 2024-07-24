package main

import (
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
)

import (
	"context"
	"errors"
	pb "testgrpc/messenger"
)

type ExistBool struct {
	Exist bool `json:"exist"`
}

type pair struct {
	first  int32
	second int32
}

type messengerServer struct {
	pb.UnimplementedMessengerServer
	savedUsers   []*pb.User
	mu           sync.Mutex
	messageCount int
	messageList  map[pair][]*pb.Message
}

func (s *messengerServer) AddUser(ctx context.Context, userInfo *pb.AddUserRequest) (*pb.UserID, error) {
	if len(userInfo.GetUsername()) < 3 {
		return &pb.UserID{}, errors.New("invalid username")
	}
	for _, user := range s.savedUsers {
		if user.GetUsername() == userInfo.GetUsername() {
			return &pb.UserID{}, errors.New("invalid username")
		}
	}
	requestUrl := "http://127.0.0.1:8000/existfile?fileID=" + userInfo.GetProfileFileID()
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return &pb.UserID{}, err
	}
	res, err := http.DefaultClient.Do(req)
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &pb.UserID{}, err
	}
	var exist ExistBool
	json.Unmarshal(resBody, &exist)
	if exist.Exist == false {
		return &pb.UserID{}, errors.New("file does not exist")
	}
	var id int32
	if len(s.savedUsers) == 0 {
		id = 1
	} else {
		id = int32(len(s.savedUsers) + 1)
	}
	userID := pb.UserID{ID: &id}
	username := userInfo.GetUsername()
	profileFileID := userInfo.GetProfileFileID()
	newUser := pb.User{ID: &userID, Username: &username, ProfileFileID: &profileFileID}
	s.savedUsers = append(s.savedUsers, &newUser)
	return &userID, nil
}

func (s *messengerServer) SendMessage(ctx context.Context, sendMessageRequest *pb.SendMessageRequest) (*pb.Empty, error) {
	haveUser1 := false
	haveUser2 := false
	for _, user := range s.savedUsers {
		if user.GetID().GetID() == sendMessageRequest.GetUserID1().GetID() {
			haveUser1 = true
		}
		if user.GetID().GetID() == sendMessageRequest.GetUserID1().GetID() {
			haveUser2 = true
		}
	}
	if haveUser1 == false {
		return &pb.Empty{}, errors.New("sender does not exists")
	}
	if haveUser2 == false {
		return &pb.Empty{}, errors.New("receiver does not exists")
	}
	if sendMessageRequest.GetContent().GetType() == pb.Content_FILE {
		requestUrl := "http://127.0.0.1:8000/existfile?fileID=" + sendMessageRequest.GetContent().GetContent()
		req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
		if err != nil {
			return &pb.Empty{}, err
		}
		res, err := http.DefaultClient.Do(req)
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return &pb.Empty{}, err
		}
		var exist ExistBool
		json.Unmarshal(resBody, &exist)
		if exist.Exist == false {
			return &pb.Empty{}, errors.New("file does not exist")
		}
	}
	var pairOfUsers pair = pair{sendMessageRequest.GetUserID1().GetID(), sendMessageRequest.GetUserID2().GetID()}
	s.messageCount += 1
	var id = int32(s.messageCount)
	messageID := &pb.MessageID{ID: &id}
	message := &pb.Message{Content: sendMessageRequest.GetContent(), ID: messageID}
	s.messageList[pairOfUsers] = append(s.messageList[pairOfUsers], message)
	return &pb.Empty{}, nil
}

func (s *messengerServer) FetchMessage(ctx context.Context, messageID *pb.MessageID) (*pb.Message, error) {
	for _, messages := range s.messageList {
		for _, message := range messages {
			if message.GetID().GetID() == messageID.GetID() {
				return message, nil
			}
		}
	}
	return &pb.Message{}, errors.New("message does not exist")
}

func (s *messengerServer) GetUserMessage(ctx context.Context, getUserMessagesRequest *pb.GetUserMessageRequest) (*pb.Chats, error) {
	chatsContent := make([]*pb.Chat, 0)
	for pairOfUser, messages := range s.messageList {
		if pairOfUser.first == getUserMessagesRequest.GetUserID().GetID() {
			userID := &pb.UserID{ID: &pairOfUser.second}
			chatContent := make([]*pb.Message, 0)
			for _, message := range messages {
				chatContent = append(chatContent, message)
			}
			chat := &pb.Chat{UserID: userID, Messages: chatContent}
			chatsContent = append(chatsContent, chat)
		}
	}
	chats := &pb.Chats{Chats: chatsContent}
	return chats, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	s := &messengerServer{messageList: make(map[pair][]*pb.Message), savedUsers: make([]*pb.User, 0), messageCount: 0}
	pb.RegisterMessengerServer(grpcServer, s)
	grpcServer.Serve(lis)
}
