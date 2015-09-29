package main

import (
  "io"
  "time"
  "flag"
  "fmt"
  "google.golang.org/grpc"
  "golang.org/x/net/context"
  pb "./protos/"
  "math/rand"
  "sort"
)

var (
  serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
  letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func randSeq(n int) string {
  b := make([]rune, n)
  for i := range b {
    b[i] = letters[rand.Intn(len(letters))]
  }
  return string(b)
}

func getInt(client pb.PerfServiceClient, intMessage *pb.IntMessage) (int32){
  intReply, err := client.SendInt(context.Background(), intMessage)
  if err != nil {
    fmt.Printf("Failed getting int!\n")
  }
  return intReply.Replynumber
}

func getDouble(client pb.PerfServiceClient, doubleMessage *pb.DoubleMessage) (float64) {
  doubleReply, err := client.SendDouble(context.Background(), doubleMessage)
  if err != nil {
    fmt.Printf("Failed getting double!\n")
  }
  return doubleReply.Replynumber
}

func getString(client pb.PerfServiceClient, stringMessage *pb.StringMessage) (string) {
  stringReply, err := client.SendString(context.Background(), stringMessage)
  if err != nil {
    fmt.Printf("Failed getting string!\n")
  }
  return stringReply.Stringreply
}

func listColleges(client pb.PerfServiceClient, stringMessage *pb.StringMessage) {
  totalBytes := 0
  t := time.Now().UnixNano()
  stream, err := client.ListColleges(context.Background(), stringMessage)
  if err != nil {
    fmt.Printf("Failed server streaming!\n")
  }
  for {
    college, err := stream.Recv()
    if err == io.EOF {
      break
    }
    if err != nil {
      fmt.Printf("Failed getting college while server streaming\n")
    }
    //fmt.Printf(college.Stringreply)
    //fmt.Printf("\n")
    totalBytes += len(college.Stringreply)
  }
  t2 := time.Now().UnixNano()
  diff := getDiff(t, t2)
  megabytes := float64(totalBytes) / float64(1000000)
  seconds := float64(diff) / float64(1000000000)
  bandwidth := megabytes / seconds
  //fmt.Printf("Total Bytes: %v Nanoseconds: %v\n", totalBytes, diff)
  fmt.Printf("Megabytes: %v Seconds: %v\n", megabytes, seconds)
  fmt.Printf("Bandwidth: %v megabytes / second\n", bandwidth)
}

func sendPreferences(client pb.PerfServiceClient) {
  totalBytes := 0
  t := time.Now().UnixNano()
  stream, err := client.SendPreferences(context.Background())
  if err != nil {
    fmt.Printf("Failed client streaming!\n")
  }
  for i := 0; i < 100; i++ {
    exampleString := randSeq(1000000)
    if err := stream.Send(&pb.StringMessage{exampleString}); err != nil {
      fmt.Printf("Failed sending preferences while client streaming\n")
    }
    totalBytes += len(exampleString)
  }
  reply, err := stream.CloseAndRecv()
  t2 := time.Now().UnixNano()
  diff := getDiff(t, t2)
  if err != nil {
    fmt.Printf("Failed while closing client stream\n")
  }
  fmt.Printf("Client streaming Reply: %v\n", reply.Stringreply)
  megabytes := float64(totalBytes) / float64(1000000)
  seconds := float64(diff) / float64(1000000000)
  bandwidth := megabytes / seconds
  fmt.Printf("Megabytes: %v Seconds: %v\n", megabytes, seconds)
  fmt.Printf("Bandwidth: %v\n", bandwidth)
}

func getDiff(t int64, t2 int64) (int64) {
  return t2 - t
}

func packInts(trials int) ([]int64) {
  l := make([]int64, trials)
  for i := 0; i < trials; i++ {
    t := time.Now().UnixNano()
    intMessage := &pb.IntMessage {int32(rand.Int())}
    t2 := time.Now().UnixNano()
    diff := getDiff(t, t2)
    l[i] = diff
    //fmt.Printf("Time to pack IntMessage: %v nanoseconds\n", diff)
    fmt.Printf("PackIntMessage: %v\n", intMessage)
  }
  return l
}

func packDoubles(trials int) ([]int64) {
  l := make([]int64, trials)
  for i := 0; i < trials; i++ {
    t := time.Now().UnixNano()
    doubleMessage := &pb.DoubleMessage {rand.Float64()}
    t2 := time.Now().UnixNano()
    diff := getDiff(t, t2)
    l[i] = diff
    //fmt.Printf("Time to pack IntMessage: %v nanoseconds\n", diff)
    fmt.Printf("PackDoubleMessage: %v\n", doubleMessage)
  }
  return l
}

func packStrings(trials int, size int) ([]int64) {
  l := make([]int64, trials)
  for i := 0; i < trials; i++ {
    t := time.Now().UnixNano()
    stringMessage := &pb.StringMessage {randSeq(size)}
    t2 := time.Now().UnixNano()
    diff := getDiff(t, t2)
    l[i] = diff
    //fmt.Printf("Time to pack IntMessage: %v nanoseconds\n", diff)
    fmt.Printf("PackStringMessage: %v\n", stringMessage)
  }
  return l
}

type int64arr []int64
func (a int64arr) Len() int { return len(a)}
func (a int64arr) Swap(i, j int) {a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool {return a[i] < a[j] }

func getMedian(measurements []int64) (float64) {
  sort.Sort(int64arr(measurements))
  if (len(measurements) % 2 == 0) {
    halfIndex := len(measurements) / 2
    return float64(measurements[halfIndex - 1] + measurements[halfIndex]) / float64(2)
  }
  return float64(measurements[len(measurements) / 2])
}

func getAverage(measurements []int64) (float64) {
  var sum int64
  sum = 0
  length := len(measurements)
  for i := 0; i < length; i++ {
    sum += measurements[i]
  }
  return float64(sum) / float64(length)
}

func main() {
  fmt.Printf("Starting client...\n") 
  var opts []grpc.DialOption
  opts = append(opts, grpc.WithInsecure())
  conn, err := grpc.Dial(*serverAddr, opts...)
  if err != nil {
    fmt.Printf("Fail to dial: %v\n", err)
  }
  defer conn.Close()
  client := pb.NewPerfServiceClient(conn)

  /*
  trials := 100
  intMeasurements := packInts(trials)
  doubleMeasurements := packDoubles(trials)
  stringMeasurements := packStrings(trials, 10)

  fmt.Printf("Int Packing Average: %v nanoseconds.\n", getAverage(intMeasurements))
  fmt.Printf("Int Packing Median: %v nanoseconds.\n", getMedian(intMeasurements))
  fmt.Printf("Double Packing Average: %v nanoseconds.\n", getAverage(doubleMeasurements))
  fmt.Printf("Double Packing Median: %v nanoseconds.\n", getMedian(doubleMeasurements))
  fmt.Printf("String Packing Average: %v nanoseconds.\n", getAverage(stringMeasurements))
  fmt.Printf("String Packing Median: %v nanoseconds.\n", getMedian(stringMeasurements))
  */

  for i := 0; i < 5; i++ {
    t := time.Now().UnixNano()
    value := getInt(client, &pb.IntMessage{int32(rand.Int())})
    t2 := time.Now().UnixNano()
    diff := getDiff(t, t2)
    fmt.Printf("RTT for IntMessage: %v nanoseconds\n", diff)
    fmt.Printf("Int Returned value: %v\n", value)
  }

  fmt.Printf("\n\n")
  t := time.Now().UnixNano()
  intMessage := &pb.IntMessage {17}
  t2 := time.Now().UnixNano()
  diff := getDiff(t, t2)
  fmt.Printf("Time to pack IntMessage: %v nanoseconds\n", diff)

  t = time.Now().UnixNano()
  value := getInt(client, intMessage)
  t2 = time.Now().UnixNano()
  diff = getDiff(t, t2)
  fmt.Printf("RTT for IntMessage: %v nanoseconds\n", diff)
  fmt.Printf("Int Returned value: %v\n", value)

  t = time.Now().UnixNano()
  doubleMessage := &pb.DoubleMessage{3.0}
  t2 = time.Now().UnixNano()
  diff = getDiff(t, t2)
  fmt.Printf("Time to pack DoubleMessage: %v nanoseconds\n", diff)

  t = time.Now().UnixNano()
  doubleValue := getDouble(client, doubleMessage)
  t2 = time.Now().UnixNano()
  diff = getDiff(t, t2)
  fmt.Printf("RTT for DoubleMessage: %v nanoseconds\n", diff)
  fmt.Printf("Double Returned value: %v\n", doubleValue)

  exampleString := randSeq(10)
  t = time.Now().UnixNano()
  stringMessage := &pb.StringMessage{exampleString}
  t2 = time.Now().UnixNano()
  diff = getDiff(t, t2)
  fmt.Printf("Time to pack StringMessage of size %v bytes: %v nanoseconds\n", len(exampleString), diff)

  t = time.Now().UnixNano()
  stringValue := getString(client, stringMessage)
  t2 = time.Now().UnixNano()
  diff = getDiff(t, t2)
  fmt.Printf("RTT for StringMessage: %v nanoseconds\n", diff)
  fmt.Printf("String returned: %v\n", stringValue)

  listColleges(client, &pb.StringMessage{"Send me random bytes!\n"})

  sendPreferences(client)
}
