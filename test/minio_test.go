package test

import (
	"context"
	"fmt"
	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"testing"
)

func InitMinIO() *minio.Client {
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds: credentials.NewStaticV4("IUID8yOPM25XR7EnOXLV", "mqPbkZd0Oj9pFtAztJ1m54YEMGrrOqnS9uzHYske", ""),
	})
	if err != nil {
		panic(err)
	}

	return client
}

func TestGetObject(t *testing.T) {
	client := InitMinIO()

	uid := 67275584949981184
	res := make([]model.PartStatusResp, 0, 100)
	info := client.ListObjects(context.Background(), "cloud-file", minio.ListObjectsOptions{
		Prefix:    fmt.Sprintf("%d/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248", uid),
		Recursive: true,
	})
	for i := range info {
		a := model.PartStatusResp{
			ObjectKey: i.Key,
			ETag:      i.ETag,
		}
		res = append(res, a)
		fmt.Println(a)
	}
}

//{
//    "code": 20000,
//    "message": "ok",
//    "data": {
//        "uploadId": "67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248",
//        "chunkUrls": [
//            {
//                "partNumber": 1,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/1?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=7942a1374b81a04ea36707142195de375e85682b238b454e2edace6e51db089a"
//            },
//            {
//                "partNumber": 2,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/2?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=2d28c171c6b1fbfbb8bacca55e9782df310879e5b6d41656208552e5dad48bbe"
//            },
//            {
//                "partNumber": 3,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/3?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=40876c596367aab7d0bf5311c2e1a975ecacc171939ad12780a032781b81b03e"
//            },
//            {
//                "partNumber": 4,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/4?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=f17103443f4f5e5069bfabdbf125f0cb48c97f70ff51a0a9c2dacee13a0a7985"
//            },
//            {
//                "partNumber": 5,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/5?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=ca3830f71f0956d514076166d3565a927380d6fa3f25c1bc5b04a91f0c56191d"
//            },
//            {
//                "partNumber": 6,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/6?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=32d301c08ccd85c183a2f412b975e5260ef9d929a1643c177156526e9ea0b7be"
//            },
//            {
//                "partNumber": 7,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/7?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=9c465b8d0bce779d9e4a7848f397985b5c8be94f32e9e3879e5f4e16376495c4"
//            },
//            {
//                "partNumber": 8,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/8?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b620237625c087e8bbb508ffad58c10ba83eafb617cc73a981b850da5e60c9a0"
//            },
//            {
//                "partNumber": 9,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/9?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=3965448a4a5877fae7a09a027e8aedf22dbd09452c2e4af15fe262e360a34bac"
//            },
//            {
//                "partNumber": 10,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/10?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=39bb4a33374e22b432ba5daa7f2a3e90974acf1d211ae5912f3f2b97efd65f2b"
//            },
//            {
//                "partNumber": 11,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/11?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=4fca00e2cffe3644bc53fa04d53da4018a597a86059ee074d5b7d8e8c7e3bf2c"
//            },
//            {
//                "partNumber": 12,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/12?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=ef52c1b330949afd446e3c6eb5dd0dc8d62a8814c1a9eb3a1db03ad6b9234639"
//            },
//            {
//                "partNumber": 13,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/13?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=cd97267f015138f3510026cd42330beba6f4533e2f1710918186c7376c50f041"
//            },
//            {
//                "partNumber": 14,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/14?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=85f23168dc36d3280dda94d33354d6c84672a118844708bb379a1e8cdb68d1fe"
//            },
//            {
//                "partNumber": 15,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/15?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=d4f3c5b39f52c38e8a531820d46390a5272821ccdda73a613f83e8497e4c76c9"
//            },
//            {
//                "partNumber": 16,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/16?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=6d34d25c655dc355f814c0e8825f33fd187144964524d474d3b9688e48421140"
//            },
//            {
//                "partNumber": 17,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/17?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b18bc00f4b21aea3ac82a14284d233108108676d106c99637de90fbe406cbbbd"
//            },
//            {
//                "partNumber": 18,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/18?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=2a76ff5f63fc6ee3703bb9fbb72f41cce20221ca147e9c21acf8209da97e546e"
//            },
//            {
//                "partNumber": 19,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/19?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=e7f440d6fbad08372de615a7fd231ae15fd0449607a39b02c43fe36d89aa93ea"
//            },
//            {
//                "partNumber": 20,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/20?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=4e031f8426e6d23fc89e463194ce387c6e80b4bff0afb0a8303592eeebbc3d3c"
//            },
//            {
//                "partNumber": 21,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/21?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b77f6c9eb238f2b628d9dbb53514b64e2cd0edb79df3bcd4d73303e6fb5b5a44"
//            },
//            {
//                "partNumber": 22,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/22?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b86d88fee775bfd39cfe91666f55f76f63ed283cec9d18fd6302881bc94c6374"
//            },
//            {
//                "partNumber": 23,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/23?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=45b0a5ab1624d737175885b2cf1ad22fdd8a055b16879b2d268f469dbe9e0af7"
//            },
//            {
//                "partNumber": 24,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/24?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=71c71a134aa61e9d48a008a6ca5224cf4baedf0ab4665406436d8e6addc519d0"
//            },
//            {
//                "partNumber": 25,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/25?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=04d1fba5e9c97247963193a459fabb0ecc7a472e110244a5c8bf7e75112e6d01"
//            },
//            {
//                "partNumber": 26,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/26?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=271c1b34d3f46e115ba09227f723eff3e192e04cd37caccf12208469f2457ec6"
//            },
//            {
//                "partNumber": 27,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/27?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=49f2a982d64806e0076fe1284b169a2b9ec054b49cc06fa23d2a2baf348390fd"
//            },
//            {
//                "partNumber": 28,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/28?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=2d204898a8fecf668bdbebaa286cbc9ecce9fee6083c54fec679a1c9b22546fe"
//            },
//            {
//                "partNumber": 29,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/29?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=9dca1a37bff98fb2c53e52707c63b21cfc74b3b1e1f3902e6485adf28be76b1c"
//            },
//            {
//                "partNumber": 30,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/30?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=e15bccff015b89278f2ad792222fc636c43ab988335d05d7c6d4616035bcb4d9"
//            },
//            {
//                "partNumber": 31,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/31?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=109c98e8f88cf66459f012a52111172f5671c40e5cf72e7acd1062ee84b53357"
//            },
//            {
//                "partNumber": 32,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/32?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=2ad4acd62d081eac56ccc3237f6ab8c9966c4f68e0dae037511dfc8b6c0a31d9"
//            },
//            {
//                "partNumber": 33,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/33?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=0489cdb5ce926df524389d706b2ac80e201e71c77a02e6d0a5447b2e4d7b17a2"
//            },
//            {
//                "partNumber": 34,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/34?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=4bd0c922e0b0b92abd578e49053c65e6bb9c3359ec757b9a9dec25b523c4f707"
//            },
//            {
//                "partNumber": 35,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/35?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=2bcc143ee5ba4ed79714d82bd8d7e3974be0ebe104497c091f2061ff53df08e8"
//            },
//            {
//                "partNumber": 36,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/36?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=e51ab4679be0ea6bbafd950e3e92256b497f73166e4672ae2a34286f38764ab0"
//            },
//            {
//                "partNumber": 37,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/37?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=fc192181f26964c7aeaf5f5abb056d25e6717e26ffe49bfb93711d26c3e44ad2"
//            },
//            {
//                "partNumber": 38,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/38?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=492618d35859aeb4a4fd4676742cb47c1d4d07d185378e44baa616e7777ca495"
//            },
//            {
//                "partNumber": 39,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/39?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=72712d19f6c962a247e50a69fc584a078fafc3aba2d6a5e710f0143f4bddd694"
//            },
//            {
//                "partNumber": 40,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/40?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=5ce90800b34fabb1f9cfd39d6c0ff0ee2c171dc52dee87fdf5801a74c635dd6c"
//            },
//            {
//                "partNumber": 41,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/41?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=8395cdfdedd2c3e6966da3c13f4f9b37f7b67f6d2a4137f975e341efb734f290"
//            },
//            {
//                "partNumber": 42,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/42?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=82ffe7556601bf5a00b4d41836db48632ff52f179d68013430c041b220ea904d"
//            },
//            {
//                "partNumber": 43,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/43?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=19aaf080d1aaec4db15c4f798108b470ad5867ab6e860f547977ac3a705668ea"
//            },
//            {
//                "partNumber": 44,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/44?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=58a87ea73915cee3214d6dde374130ce9c1f112925b17d942181dc42987cc63b"
//            },
//            {
//                "partNumber": 45,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/45?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=c4f4354f457f299b9ad1f641ec821d4520be9df069ae92176b38a9ebba894085"
//            },
//            {
//                "partNumber": 46,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/46?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b54fda304a91fd1b83eaa4c8b75ac124dd6a055ee434c1cc97c141fe15f6f06c"
//            },
//            {
//                "partNumber": 47,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/47?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=0fea16548f54f2c73c0d7ed9a92e5647333b79d8897ba4aec09e5353d04f20f1"
//            },
//            {
//                "partNumber": 48,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/48?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=45b6a16811b76f854a25a49dbdf9f318badb5e6de2e2c0460d3e46235b38d983"
//            },
//            {
//                "partNumber": 49,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/49?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=e17365f33bdc13686964152c9bbb5727fefca4ce170b343c3d1f31e5cd019fd9"
//            },
//            {
//                "partNumber": 50,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/50?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=e609bfed9dd19b0bd54e19ed11671f6f0b936d97efd417b83e9afad18065702f"
//            },
//            {
//                "partNumber": 51,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/51?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=832e464e8fcf6eec2590ee807194abb1ee936137f79986fdec13a3fff6614ff4"
//            },
//            {
//                "partNumber": 52,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/52?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=0b91c000d5f96e671de8be34c04fe39ebcd603f21743af33b9482b329305df01"
//            },
//            {
//                "partNumber": 53,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/53?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=d26df2f0ac713c1cbcba75a32afed215b7ca189fb12d2101490d49752a281ef3"
//            },
//            {
//                "partNumber": 54,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/54?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=9b4fcb64cab62d5d87d55f5364bc700f9a4ce1337478bfb073199c2e19012bad"
//            },
//            {
//                "partNumber": 55,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/55?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=5ffeaa638e98b26cdf435b4a6f731a45764af2663531a0d4a04932fed1bc1aac"
//            },
//            {
//                "partNumber": 56,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/56?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=4cabc370063337d3e2f184e8053f2c0a9198d844ffbabd4b05c64db0a1188ff1"
//            },
//            {
//                "partNumber": 57,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/57?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=c6cfaa33618251f591d2e9ba24cf41075edec74e768f9112a8dacda7d7d7c271"
//            },
//            {
//                "partNumber": 58,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/58?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=698c9ede3af19f8a81e3481bedf739c15b37077dc76bdf89849ffa00e7630a70"
//            },
//            {
//                "partNumber": 59,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/59?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=5ddf03559f5a67e318ade97c44dfadbca90e6f1399e16d114d742bb0bd8f82b6"
//            },
//            {
//                "partNumber": 60,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/60?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=43c9226433eb933dd199cc20ab0b613bc8ce269c6a562109c3add17fb23f8d89"
//            },
//            {
//                "partNumber": 61,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/61?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=a3883972c9ba75f8a839dab5c5e7b08df677bee5a185881798aca239f0b8a09d"
//            },
//            {
//                "partNumber": 62,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/62?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=43dbca183db4c92c47e3d6ab80f4bbe02b9750a7e9be95f3889785b1b5900075"
//            },
//            {
//                "partNumber": 63,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/63?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=0790020c349d7f88a8547fe53145083839ef0bd35675f06e2e545e0f616b757c"
//            },
//            {
//                "partNumber": 64,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/64?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=bcdfe63cb94f02baeb0c6e3da2c5813ca032678bc36c342cb92e9cd742f35b90"
//            },
//            {
//                "partNumber": 65,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/65?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=8b305d49499add8736482443e81547821b9655a49cf197de59276304fbdc4cf1"
//            },
//            {
//                "partNumber": 66,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/66?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b4c87f81eb810759a0bb7ced75fe7eb5f8e45f9294482a49f348fc3fbc9865aa"
//            },
//            {
//                "partNumber": 67,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/67?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=c44b61e61160db7864676dffad6301465e0e1aebd26b7a19e63f6d4bd5b4557d"
//            },
//            {
//                "partNumber": 68,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/68?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=56e8dd1c91d0f8ffdde272bdf9de61285155c0ddf9999a20e306f1ddb811daa9"
//            },
//            {
//                "partNumber": 69,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/69?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=119fd6a92a5030cdb05930e574f3722893152d04e3aa8ea949697605e723176c"
//            },
//            {
//                "partNumber": 70,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/70?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=a6df9d390eac61ecd6cb8d355dd808f64975512b9c4caf82c73546a2d2016428"
//            },
//            {
//                "partNumber": 71,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/71?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=800451793f4f305a42f881fd17538b55f659e4c33f39cc7661e0c3411f9a67ac"
//            },
//            {
//                "partNumber": 72,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/72?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=1bc05008d119694804d478f2d8d8e9620645b82d696e4387e5faee4f66f50389"
//            },
//            {
//                "partNumber": 73,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/73?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=72a86bb6b727fc40547145462b60bf79ed039a9de89b0fa5fbe2e7a404d73c94"
//            },
//            {
//                "partNumber": 74,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/74?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=6d0cc3ba383c3f5a448177fed1a651c388e323f9b72cb22e3f9d5b3f5a70ee5f"
//            },
//            {
//                "partNumber": 75,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/75?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=0c4352d831368da6ca6f88d50f67549c6a46c479ac39ca33e379bc5cef128d12"
//            },
//            {
//                "partNumber": 76,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/76?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=8f61bb1fc7d781efa050bc9a3cad4fc0220f977d8933638957b2f84519b50597"
//            },
//            {
//                "partNumber": 77,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/77?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=f82b930ce48501330e7dcfa3983204a38cd16388f2a045824780e044d45b8a71"
//            },
//            {
//                "partNumber": 78,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/78?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=3844df6805d82bdf85865b3214a47a46eac5529f9e78ec9ea09743a0457af30b"
//            },
//            {
//                "partNumber": 79,
//                "presignedUrl": "http://localhost:9000/cloud-file/67275584949981184/chunks/67275584949981184_74eeb2b8f4bbc3dc91a699a63f75461035214d3e4dbf04ee07232c3f2c2396ef-822687248-1736311551000_822687248/79?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=IUID8yOPM25XR7EnOXLV%2F20250604%2Fus-east-1%2Fs3%2Faws4_request\u0026X-Amz-Date=20250604T052605Z\u0026X-Amz-Expires=3600\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=471881afa73d732665c905814fdc1a29cd905b79c16443ee9cfb87416946ca91"
//            }
//        ],
//        "expiresIn": 3600,
//        "recommendedConcurrency": 6,
//        "optimalChunkSize": 10485760,
//        "totalChunks": 79,
//        "uploadMethod": "direct-to-storage",
//        "fileExists": false
//    }
//}
