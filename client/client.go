package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pb "testgrpc/messenger"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:8080", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewMessengerClient(conn)

	username1 := "shkk8x4"
	profileFileID1 := "17126196425100764312-nnoinnlffpnnwjgbnrtnknqnbntnkxon"
	userId1, err := client.AddUser(context.Background(), &pb.AddUserRequest{Username: &username1, ProfileFileID: &profileFileID1})
	if err != nil {
		panic(err)
	}
	fmt.Println("User id:")
	fmt.Println(userId1.GetID())

	username2 := "sh84"
	profileFileID2 := "17126196425100764312-nnoinnlffpnnwjgbnrtnknqnbntnkxon"
	userId2, err := client.AddUser(context.Background(), &pb.AddUserRequest{Username: &username2, ProfileFileID: &profileFileID2})
	if err != nil {
		panic(err)
	}
	fmt.Println("User id:")
	fmt.Println(userId2.GetID())

	fileId := "17126196425100764312-nnoinnlffpnnwjgbnrtnknqnbntnkxon"
	contentType := pb.Content_FILE
	content := &pb.Content{Content: &fileId, Type: &contentType}
	_, err = client.SendMessage(context.Background(), &pb.SendMessageRequest{UserID1: userId1, UserID2: userId2, Content: content})
	if err != nil {
		panic(err)
	}

	messageContent1 := "salam khobi?"
	contentType2 := pb.Content_TEXT
	content2 := &pb.Content{Content: &messageContent1, Type: &contentType2}
	_, err = client.SendMessage(context.Background(), &pb.SendMessageRequest{UserID1: userId1, UserID2: userId2, Content: content2})
	if err != nil {
		panic(err)
	}

	messageContent2 := "salam. are. che khabar?"
	contentType3 := pb.Content_IMAGE
	content3 := &pb.Content{Content: &messageContent2, Type: &contentType3}
	_, err = client.SendMessage(context.Background(), &pb.SendMessageRequest{UserID1: userId2, UserID2: userId1, Content: content3})
	if err != nil {
		panic(err)
	}

	var id1 int32 = 1
	messageId1 := &pb.MessageID{ID: &id1}
	message1, err := client.FetchMessage(context.Background(), messageId1)
	if err != nil {
		panic(err)
	}
	fmt.Println("Message content")
	fmt.Println(message1.GetContent().GetContent())

	var id2 int32 = 3
	messageId2 := &pb.MessageID{ID: &id2}
	message2, err := client.FetchMessage(context.Background(), messageId2)
	if err != nil {
		panic(err)
	}
	fmt.Println("Message content")
	fmt.Println(message2.GetContent().GetContent())

	getUserMessageRequest := &pb.GetUserMessageRequest{UserID: userId1}
	chats1, err := client.GetUserMessage(context.Background(), getUserMessageRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println("all messages from user1")
	for _, chat := range chats1.GetChats() {
		fmt.Println(chat.GetUserID().GetID())
		for _, message := range chat.GetMessages() {
			fmt.Println(message.GetContent())
		}
	}

	getUserMessageRequest2 := &pb.GetUserMessageRequest{UserID: userId2}
	chats2, err := client.GetUserMessage(context.Background(), getUserMessageRequest2)
	if err != nil {
		panic(err)
	}
	fmt.Println("all messages from user2")
	for _, chat := range chats2.GetChats() {
		fmt.Println(chat.GetUserID().GetID())
		for _, message := range chat.GetMessages() {
			fmt.Println(message.GetContent())
		}
	}
}
