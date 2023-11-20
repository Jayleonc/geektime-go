// Package wrk
package wrk

//  wrk -t1 -d1s -c2 -s ./script/wrk/signup.lua http://localhost:8080/users/signup
// -t：线程数量
// -d：持续时间，比如 1s 是一秒，1m 是一分钟
// -c：并发数
// -s：测试脚本
