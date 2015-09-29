package main

import (
  "io"
  "fmt"
  "flag"
  "net"
  "bytes"
  "math/rand"

  "golang.org/x/net/context"
  "google.golang.org/grpc"
  pb "./protos/"
)

var (
  port = flag.Int("port", 10000, "The server port")
  letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

type serverStruct struct {
  savedInts []*pb.IntReply
}

func (s *serverStruct) SendInt(ctx context.Context, intMessage *pb.IntMessage) (*pb.IntReply, error) {
  return &pb.IntReply{intMessage.Number * 2}, nil
}

func (s *serverStruct) SendDouble(ctx context.Context, doubleMessage *pb.DoubleMessage) (*pb.DoubleReply, error) {
  return &pb.DoubleReply{doubleMessage.Number * 3.0}, nil
}

func (s *serverStruct) ListColleges(stringMessage *pb.StringMessage, stream pb.PerfService_ListCollegesServer) error {
  for i := 0; i < 100; i++ {
    returnString := randSeq(1000000)
    stringReply := &pb.StringReply{returnString}
    if err := stream.Send(stringReply); err != nil {
      return err
    }
  }
  return nil
}

func (s *serverStruct) SendPreferences(stream pb.PerfService_SendPreferencesServer) error {
  for {
    stringMessage, err := stream.Recv()
    if err == io.EOF {
      return stream.SendAndClose(&pb.StringReply{"Received all preferences!"})
    }
    if err != nil {
      return err
    }
    fmt.Printf("Message Received: %v\n", stringMessage.Stringmessage)
  }
  return nil
}

func (s *serverStruct) SendString(ctx context.Context, stringMessage *pb.StringMessage) (*pb.StringReply, error) {
  var buffer bytes.Buffer
  buffer.WriteString("Message was: ")
  buffer.WriteString(stringMessage.Stringmessage)
  return &pb.StringReply{buffer.String()}, nil
}

func randSeq(n int) string {
  b := make([]rune, n)
  for i := range b {
    b[i] = letters[rand.Intn(len(letters))]
  }
  return string(b)
}


func newServer() *serverStruct {
  s := new(serverStruct)
  return s
}

func main() {
  fmt.Printf("Starting Server\n")
  lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
  if err != nil {
    fmt.Printf("Failed to listen: %v\n", err)
  }

  var opts []grpc.ServerOption
  grpcServer := grpc.NewServer(opts...)
  pb.RegisterPerfServiceServer(grpcServer, newServer())
  grpcServer.Serve(lis) 
}
