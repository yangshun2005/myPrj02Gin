package main

import (
	"flag"
	template2 "html/template"

	"github.com/yangshun2005/myPrj02Gin/controllers"
	"github.com/yangshun2005/myPrj02Gin/helpers"
	"github.com/yangshun2005/myPrj02Gin/models"
	"github.com/yangshun2005/myPrj02Gin/system"
	"path/filepath"
	"text/template"

	"github.com/claudiu/gocron"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/cihub/seelog"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func main() {

	configFilePath := flag.String("C", "conf/conf.yaml", "config file path")

	//加载程序配置文件和日子配置文件，并检查web服务启动正常否
	logConfigPath := flag.String("L", "conf/seelog.xml", "log config file path")
	flag.Parse()

	if err := system.LoadConfiguration(*configFilePath); err != nil {
		seelog.Critical("err parsing config log file", err)
		return
	}

	logger, err := seelog.LoggerFromConfigAsFile(*logConfigPath)
	if err != nil {
		seelog.Critical("err parsing seelog config file", err)
		return
	}
	seelog.ReplaceLogger(logger)
	defer seelog.Flush()

	//设置gin框架的使用模式
	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)

	//获取gin服务指针
	route := gin.Default()

	//route.GET("/", func(c *gin.Context) {
	//	name := c.Query("name")
	//	c.JSON(http.StatusOK, gin.H{"query": name})
	//	logger.Debug()
	//})

	//连接数据库并创建tables，并返回gorm.db指针
	db, err := models.InitDB()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer db.Close()

	//设置route
	setTemplate(route)
	setSessions(route)
	route.Use(SharedData())

	//循环任务 tasks
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Every(7).Days().Do(controllers.Backup)
	gocron.Start()

	//设置静态文件目录 并使用getCurrentDirectory()预留静态文件其他目录位置
	route.Static("/static", filepath.Join(getCurrentDirectory(), "./static"))

	//设置没有uri的时候404的响应问题
	route.NoRoute(controllers.Handle404)

	//路由开发三块：用户前端、管理后台、后端逻辑、oauth认证
	//用户前端
	route.GET("/", controllers.IndexGet)
	route.GET("/index", controllers.IndexGet)
	route.GET("/rss", controllers.RssGet)

	if system.GetConfiguration().SignupEnabled {
		route.GET("/signup", controllers.SignupGet)
		route.POST("/signup", controllers.SignupPost)
	}




	//
	//
	//
	//

	//启动web服务
	route.Run(":9090")
}

//自定义gin模版函数
func setTemplate(engine *gin.Engine) {

	funcMap := template.FuncMap{
		"dateFormat": helpers.DateFormat,
		"substring":  helpers.Substring,
		"isOdd":      helpers.IsOdd,
		"isEven":     helpers.IsEven,
		"truncate":   helpers.Truncate,
		"add":        helpers.Add,
		"minus":      helpers.Minus,
		"listtag":    helpers.ListTag,
	}

	engine.SetFuncMap(template2.FuncMap(funcMap))
	engine.LoadHTMLGlob(filepath.Join(getCurrentDirectory(), "./views/**/*"))
}

//setSessions initializes sessions & csrf middlewares
//本处使用gin框架自身的session方式，也可以扩展成使用redis或者jwt的方式
func setSessions(router *gin.Engine) {
	config := system.GetConfiguration()
	//https://github.com/gin-gonic/contrib/tree/master/sessions
	store := cookie.NewStore([]byte(config.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"}) //Also set Secure: true if using SSL, you should though
	router.Use(sessions.Sessions("gin-session", store))
	//https://github.com/utrack/gin-csrf
	/*router.Use(csrf.Middleware(csrf.Options{
		Secret: config.SessionSecret,
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))*/
}

func getCurrentDirectory() string {
	return ""
}

//SharedData fills in common data, such as user info, etc...
func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uID := session.Get(controllers.SESSION_KEY); uID != nil {
			user, err := models.GetUser(uID)
			if err == nil {
				c.Set(controllers.CONTEXT_USER_KEY, user)
			}
		}
		if system.GetConfiguration().SignupEnabled {
			c.Set("SignupEnabled", true)
		}
		c.Next()
	}
}
