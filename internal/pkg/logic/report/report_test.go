package report

import (
	"testing"
)

func TestGetReportDetail(t *testing.T) {
	//ctx := context.Background()
	//conf := fmt.Sprintf("mongodb://%s:%s@%s/%s", "kunpeng", "hello123456", "172.17.79.88:37017", "kunpeng")
	//
	//clientOptions := options.Client().ApplyURI(conf)
	//mongoClient, err := mongo.Connect(ctx, clientOptions)
	//if err != nil {
	//	return
	//}
	//
	//rdb := redis.NewClient(&redis.Options{
	//	Addr:     "172.17.79.88:63790",
	//	Password: "mypassword",
	//	DB:       0,
	//})
	////
	//collection := mongoClient.Database("kunpeng").Collection("report_data")
	////
	//var report rao.GetReportReq
	//report.ReportID = 17
	//report.PlanId = 0
	//err, result := GetReportDetail(ctx, report, collection, rdb)
	//if err != nil {
	//	fmt.Println("err1:         ", err)
	//}
	//data, err := json.Marshal(result)
	//if err != nil {
	//	fmt.Println("err2:        ", err)
	//}
	//fmt.Println("data..................;          ", string(data))
	//fmt.Println(string(data))
	//client, _ := elastic.NewClient(
	//	elastic.SetURL("http://172.17.101.191:9200"),
	//	elastic.SetSniff(false),
	//	elastic.SetBasicAuth("elastic", "ZSrfx4R6ICa3skGBpCdf"),
	//	elastic.SetErrorLog(log.New(os.Stdout, "APP", log.Lshortfile)),
	//	elastic.SetHealthcheckInterval(30*time.Second),
	//)
	//_, _, err := client.Ping("http://172.17.101.191:9200").Do(context.Background())
	//if err != nil {
	//	panic(fmt.Sprintf("es连接失败: %s", err))
	//}
	//if err != nil {
	//	fmt.Println(err)
	//}
	//res, _ := json.Marshal(result)
	//fmt.Println(string(res))
	//log.Println(string(res))

	//
	//filter := bson.D{{"report_id", 1149}}
	//fmt.Println("lllllll        ", collection)
	//cur := collection.FindOne(context.TODO(), filter)
	//result, err := cur.DecodeBytes()
	//list, err := result.Elements()
	//for index, value := range list {
	//	fmt.Println("index         ", index, " value:           ", string(value.Value().Value))
	//}
	//fmt.Println("1111111", result, " errr:           ", err)
	//if cur == nil {
	//	debug := bson.D{{fmt.Sprintf("%d", 123), "All"}}
	//	_, err = collection.InsertOne(ctx, debug)
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrRedisFailed, err.Error())
	//		return
	//	}
	//} else {
	//	debug := bson.D{{fmt.Sprintf("%d", 123), "all"}}
	//	_, err = collection.UpdateMany(ctx, filter, debug)
	//	if err != nil {
	//		response.ErrorWithMsg(ctx, errno.ErrRedisFailed, err.Error())
	//		return
	//	}
	//}
}

func TestGetReportDebugStatus(t *testing.T) {

	//re := rao.GetReportReq{
	//	ReportID: 1149,
	////}
	//conf := fmt.Sprintf("mongodb://%s:%s@%s/%s", "kunpeng", "kYjJpU8BYvb4EJ9x", "172.17.18.255:27017", "kunpeng")
	//
	//clientOptions := options.Client().ApplyURI(conf)
	//mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	//if err != nil {
	//	fmt.Println("err:          ", err)
	//	return
	//}
	//
	//collection := mongoClient.Database("kunpeng").Collection("debug_status")
	////filter := bson.D{{"report_id", 1149}}
	////cur := collection.FindOne(context.TODO(), filter)
	//list, err := cur.DecodeBytes()
	//if err != nil {
	//	fmt.Println("err:             ", err)
	//}
	//
	//fmt.Println(list)
	//str := GetReportDebugStatus(context.TODO(), re)
	//fmt.Println("111111111111111", str)
	//req := rao.GetReportReq{
	//	ReportID: 1149,
	//}
	//ctx := context.TODO()
	//result := GetReportDebugStatus(ctx, req, collection)
	//fmt.Println("result:           ", result)
	//fmt.Println("123")
	//response.SuccessWithData(ctx, result)

}
