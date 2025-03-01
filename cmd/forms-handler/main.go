package main

import (
	"context"
	"fmt"
)

func main() {
	request := []byte(`{
		"body":"{\"testType\": \"test\",\"testName\": \"\\u0422\\u0435\\u0441\\u0442\\u043e\\u0432\\u0430\\u044f \\u0444\\u043e\\u0440\\u043c\\u0430\",\"answer\": {\"id\": 1993710697, \"uid\": \"2082590513\", \"data\": {\"couchEmail\": {\"value\": \"couch@test.com\",\"question\": {\"id\": 72238928, \"slug\": \"answer_non_profile_email_72238928\", \"options\": {\"required\": false},\"answer_type\": {\"id\": 32, \"slug\": \"answer_non_profile_email\"}}},\"clientEmail\": {\"value\": \"iamvkosarev@gmail.com\", \"question\": {\"id\": 72238928,\"slug\": \"answer_non_profile_email_72238928\", \"options\": {\"required\": false}, \"answer_type\": {\"id\": 32, \"slug\": \"answer_non_profile_email\"}}}, \"answer_boolean_72238954\": {\"value\": true, \"question\": {\"id\": 72238954, \"slug\": \"answer_boolean_72238954\", \"options\": {\"required\": false}, \"answer_type\": {\"id\": 33, \"slug\": \"answer_boolean\"}}}}, \"survey\": {\"id\": \"67c052daf47e73dd12eeeee7\"}, \"created\": \"2025-02-27T17:40:16Z\", \"cloud_uid\": \"aje4faar3ggap9889j6t\"}}"
	}`)

	response, err := Handler(context.Background(), request)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Response:", response)
}
