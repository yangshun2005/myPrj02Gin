package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
验证使用程序文件，项目提交完成即删除
 */

import (
	"os"
)

func main01() {
	var cliName = flag.String("name", "nick", "Input Your Name")

	//fmt.Fprintf(os.Stderr, *cliName)
	flag.PrintDefaults()
	fmt.Printf(*cliName + "\n")

	arg := os.Args[1:]

	if arg[1] == *cliName {
		fmt.Println("ok")
	} else {
		fmt.Println(*cliName)
	}
	flag.Parse()
	fmt.Println("输出：", arg)


}

//获取uri
func main02() {
	engine := gin.Default()
	engine.GET("/william/", func(c *gin.Context) {
		c.JSON(http.StatusOK,gin.H{"path":c.Request.URL.Path})
	})
	engine.Run(":9091")
}

//设置cokkie
func main03() {

	router := gin.Default()

	router.GET("/cookie", func(c *gin.Context) {

		cookie, err := c.Cookie("gin_cookie")

		if err != nil {
			cookie = "NotSet"
			c.SetCookie("gin_cookie", "test", 3600, "/", "localhost", false, true)
		}

		fmt.Printf("Cookie value: %s \n", cookie)
	})


	router.Run(":9091")
}


//session使用
func main04() {
	r := gin.Default()
	// 创建基于cookie的存储引擎，secret11111 参数是用于加密的密钥
	store := cookie.NewStore([]byte("secret11111"))
	// 设置session中间件，参数mysession，指的是session的名字，也是cookie的名字
	// store是前面创建的存储引擎，我们可以替换成其他存储引擎
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/hello", func(c *gin.Context) {
		// 初始化session对象
		session := sessions.Default(c)

		// 通过session.Get读取session值
		// session是键值对格式数据，因此需要通过key查询数据
		if session.Get("hello") != "world" {
			// 设置session数据
			session.Set("hello", "world")
			// 删除session数据
			session.Delete("tizi365")
			// 保存session数据
			session.Save()
			// 删除整个session
			session.Clear()
		}


		c.JSON(200, gin.H{"hello": session.Get("hello")})
	})
	r.Run(":9091")
}


//使用redis存储session
func main() {
	r := gin.Default()
	// 初始化基于redis的存储引擎
	// 参数说明：
	//    第1个参数 - redis最大的空闲连接数
	//    第2个参数 - 数通信协议tcp或者udp
	//    第3个参数 - redis地址, 格式，host:port
	//    第4个参数 - redis密码
	//    第5个参数 - session加密密钥
	store, _ := redis.NewStore(10, "tcp", "localhost:6379", "password", []byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/incr", func(c *gin.Context) {
		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		if v == nil {
			count = 0
		} else {
			count = v.(int)
			count++
		}
		session.Set("count", count)
		session.Save()
		c.JSON(200, gin.H{"count": count})
	})
	r.Run(":9091")
}