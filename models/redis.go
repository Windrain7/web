package models

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

//TODO bookCourse的同步问题还没考虑
const bookCourseScript = `
	--- KEYS[1]:courseIdKey	 "course_id_{courseId}"
	--- KEYS[2]:bookCourseKey "book_{studentId}"
	--- 返回-1表示课程已经没有余额，返回-2表示课程已经选过

	local remain = redis.call("HGET", KEYS[1], "remain");
	if (tonumber(remain) == 0) 
	then 
		return -1;
	end
	local bookCourse = redis.call("SISMEMBER", KEYS[2], KEYS[1]);
	if (bookCourse == 1)
	then 
		return -2;
	end
	redis.call("HSET", KEYS[1], "remain", remain-1);
	redis.call("SADD", KEYS[2], KEYS[1]);
	return 1;
`

var BookSHA string
var Ctx = context.Background()
var Rdb *redis.Client

func init() {

	Rdb = redis.NewClient(&redis.Options{
		Addr:     "180.184.74.66:6379",
		Password: "20220121",
		DB:       0,
	})
	pong, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Println("连接redis失败", pong, err)
	}
	Rdb.FlushDB(Ctx)
	if BookSHA, err = Rdb.ScriptLoad(Ctx, bookCourseScript).Result(); err != nil {
		log.Println("上传脚本错误")
	}

}
