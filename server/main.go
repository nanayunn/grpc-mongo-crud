package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
	_ "strings"
	"net/url"

	gw "github.com/nanayunn/grpc-mongo-crud/gateway"
	userpb "github.com/nanayunn/grpc-mongo-crud/proto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
}

type UserItem struct {
	ID     primitive.ObjectID `json: "_id" bson:"_id"`
	Name   string             `json: "name" bson:"name"`
	Age    string             `json: "age" bson:"age"`
	Userid string             `json: "userid" bson:"userid"`
}

var db *mongo.Client
var userdb *mongo.Collection
var mongoCtx context.Context

func (u *UserServiceServer) ReadUser(ctx context.Context, req *userpb.ReadUserReq) (*userpb.ReadUserRes, error) {
	// convert string id (from proto) to MongoDB Object id
	object_id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid Format Id, Cannot convert to ObjectID: %v", err))
	}
	result := userdb.FindOne(ctx, bson.M{"_id": object_id})
	data := UserItem{}

	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find user with Object Id %s : %v", req.GetId(), err))
	}
	response := &userpb.ReadUserRes{
		User: &userpb.User{
			Id:     object_id.Hex(),
			Name:   data.Name,
			Age:    data.Age,
			Userid: data.Userid,
		},
	}
	return response, nil
}

func (u *UserServiceServer) CreateUser(ctx context.Context, req *userpb.CreateUserReq) (*userpb.CreateUserRes, error) {
	user := req.GetUser()
	fmt.Println("user :  ", user)
	data := UserItem{
		ID:     primitive.NewObjectID(),
		Name:   user.GetName(),
		Age:    user.GetAge(),
		Userid: user.GetUserid(),
	}
	fmt.Println("data :  ", data)
	result, err := userdb.InsertOne(ctx, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("CreateUser : Internal Error : %v", err))
	}
	object_id := result.InsertedID.(primitive.ObjectID)
	user.Id = object_id.Hex()

	return &userpb.CreateUserRes{User: user}, nil

}

func (u *UserServiceServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserReq) (*userpb.UpdateUserRes, error) {
	user := req.GetUser()

	object_id, err := primitive.ObjectIDFromHex(user.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not get the user information with supplied MongoDB ObjectID: %v", err))
	}

	update := bson.M{
		"name":   user.GetName(),
		"age":    user.GetAge(),
		"userid": user.GetUserid(),
	}

	filter := bson.M{"_id": object_id}

	result := userdb.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := UserItem{}

	err = result.Decode(&decoded)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find user with supplied ID: %v", err))
	}

	return &userpb.UpdateUserRes{
		User: &userpb.User{
			Id:     decoded.ID.Hex(),
			Name:   decoded.Name,
			Age:    decoded.Age,
			Userid: decoded.Userid,
		},
	}, nil

}

func (u *UserServiceServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserReq) (*userpb.DeleteUserRes, error) {
	object_id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert supplied Obejct ID: %v", err),
		)
	}
	_, err = userdb.DeleteOne(ctx, bson.M{"_id": object_id})
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Could not find user with id &s: %v", object_id),
		)
	}

	return &userpb.DeleteUserRes{
		Success: true,
	}, nil
}

func (u *UserServiceServer) ListUser(req *userpb.ListUserReq, stream userpb.UserService_ListUserServer) error {
	data := &UserItem{}

	cursor, err := userdb.Find(context.Background(), bson.M{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal Error: %v", err),
		)
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Unavailable,
				fmt.Sprintf("Could not decode data : %v", err),
			)
		}
		stream.Send(&userpb.ListUserRes{
			User: &userpb.User{
				Id:     data.ID.Hex(),
				Name:   data.Name,
				Age:    data.Age,
				Userid: data.Userid,
			},
		})
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown cursor error: %v", err),
		)
	}
	return nil
}

const (
	username         = "piolink"
	password         = "piolink0726"
	hostname         = "example-mongodb-svc.default.svc.cluster.local"
	//hostname	 = "10.244.1.23"
	mongo_port       = "27017"
	server_port      = "0.0.0.0:50051"
	grpc_server_port = "0.0.0.0:50050"
)

func getConnectionURI() string {
	return "mongodb://" + url.QueryEscape(username) + ":" + url.QueryEscape(password) + "@" + hostname + ":" + mongo_port
}

func ConnectDB() (client *mongo.Client, ctx context.Context) {

	//conn_mongo := []string{hostname, mongo_port}
	//conn_mongo_add := strings.Join(conn_mongo, ":")

	conn_mongo_add := getConnectionURI()

	log.Printf("Connecting to Mongo .. URL : %s\n", conn_mongo_add)

	// Timeout 설정을 위한 Context생성
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	log.Printf("Trying to connect to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conn_mongo_add).SetAuth(options.Credential{
		Username: username,
		Password: password,
	}))

	if err != nil {
		log.Fatal(err)
	}
	if err := client.Ping(context.TODO(), nil); err != nil{
		log.Fatal(err)

	}

	log.Printf("MongoDB Connection Created..")

	return client, ctx
}

func main() {
	listener, err := net.Listen("tcp", server_port)
	if err != nil {
		log.Fatalf("Unable to listen on port %s: %v", server_port, err)
	}

	opts := []grpc.ServerOption{}

	grpc_server := grpc.NewServer(opts...)

	srv := &UserServiceServer{}

	userpb.RegisterUserServiceServer(grpc_server, srv)


	db, mongoCtx = ConnectDB()

	userdb = db.Database("mydb_user").Collection("User")

	go func() {
		if err := grpc_server.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	fmt.Println("Server Connected! Listening on port : ", server_port)

	if err := gw.RungrpcGateway(); err != nil {
		log.Fatalf("Failed to Open HTTP gw Server..")
	}

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	<-c

	fmt.Println("\nDisconnecting Server.. server stopping..")

	grpc_server.Stop()

	listener.Close()
	fmt.Println("Server Closed..")
	db.Disconnect(mongoCtx)
	fmt.Println("MongoDB Connection Closed..")

}
