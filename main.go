package main

import (
	"flag"
	template2 "html/template"
	"net/http"

	"path/filepath"
	"text/template"

	"github.com/yangshun2005/myPrj02Gin/controllers"
	"github.com/yangshun2005/myPrj02Gin/helpers"
	"github.com/yangshun2005/myPrj02Gin/models"
	"github.com/yangshun2005/myPrj02Gin/system"

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

	// user signin and logout
	route.GET("/signin", controllers.SigninGet)
	route.POST("/signin", controllers.SigninPost)
	route.GET("/logout", controllers.LogoutGet)
	route.GET("/oauth2callback", controllers.Oauth2Callback)
	route.GET("/auth/:authType", controllers.AuthGet)

	// captcha
	route.GET("/captcha", controllers.CaptchaGet)

	visitor := route.Group("/visitor")
	visitor.Use(AuthRequired())
	{
		visitor.POST("/new_comment", controllers.CommentPost)
		visitor.POST("/comment/:id/delete", controllers.CommentDelete)
	}

	// subscriber
	route.GET("/subscribe", controllers.SubscribeGet)
	route.POST("/subscribe", controllers.Subscribe)
	route.GET("/active", controllers.ActiveSubscriber)
	route.GET("/unsubscribe", controllers.UnSubscribe)

	route.GET("/page/:id", controllers.PageGet)
	route.GET("/post/:id", controllers.PostGet)
	route.GET("/tag/:tag", controllers.TagGet)
	route.GET("/archives/:year/:month", controllers.ArchiveGet)

	route.GET("/link/:id", controllers.LinkGet)

	authorized := route.Group("/admin")
	authorized.Use(AdminScopeRequired())
	{
		// index
		authorized.GET("/index", controllers.AdminIndex)

		// image upload
		authorized.POST("/upload", controllers.Upload)

		// page
		authorized.GET("/page", controllers.PageIndex)
		authorized.GET("/new_page", controllers.PageNew)
		authorized.POST("/new_page", controllers.PageCreate)
		authorized.GET("/page/:id/edit", controllers.PageEdit)
		authorized.POST("/page/:id/edit", controllers.PageUpdate)
		authorized.POST("/page/:id/publish", controllers.PagePublish)
		authorized.POST("/page/:id/delete", controllers.PageDelete)

		// post
		authorized.GET("/post", controllers.PostIndex)
		authorized.GET("/new_post", controllers.PostNew)
		authorized.POST("/new_post", controllers.PostCreate)
		authorized.GET("/post/:id/edit", controllers.PostEdit)
		authorized.POST("/post/:id/edit", controllers.PostUpdate)
		authorized.POST("/post/:id/publish", controllers.PostPublish)
		authorized.POST("/post/:id/delete", controllers.PostDelete)

		// tag
		authorized.POST("/new_tag", controllers.TagCreate)

		//
		authorized.GET("/user", controllers.UserIndex)
		authorized.POST("/user/:id/lock", controllers.UserLock)

		// profile
		authorized.GET("/profile", controllers.ProfileGet)
		authorized.POST("/profile", controllers.ProfileUpdate)
		authorized.POST("/profile/email/bind", controllers.BindEmail)
		authorized.POST("/profile/email/unbind", controllers.UnbindEmail)
		authorized.POST("/profile/github/unbind", controllers.UnbindGithub)

		// subscriber
		authorized.GET("/subscriber", controllers.SubscriberIndex)
		authorized.POST("/subscriber", controllers.SubscriberPost)

		// link
		authorized.GET("/link", controllers.LinkIndex)
		authorized.POST("/new_link", controllers.LinkCreate)
		authorized.POST("/link/:id/edit", controllers.LinkUpdate)
		authorized.POST("/link/:id/delete", controllers.LinkDelete)

		// comment
		authorized.POST("/comment/:id", controllers.CommentRead)
		authorized.POST("/read_all", controllers.CommentReadAll)

		// backup
		authorized.POST("/backup", controllers.BackupPost)
		authorized.POST("/restore", controllers.RestorePost)

		// mail
		authorized.POST("/new_mail", controllers.SendMail)
		authorized.POST("/new_batchmail", controllers.SendBatchMail)
	}

	//
	//
	//
	//

	//启动web服务
	//route.Run(":9090")
	route.Run(system.GetConfiguration().Addr)
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

//AuthRequired grants access to authenticated users, requires SharedData middleware
func AdminScopeRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(controllers.CONTEXT_USER_KEY); user != nil {
			if u, ok := user.(*models.User); ok && u.IsAdmin {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(controllers.CONTEXT_USER_KEY); user != nil {
			if _, ok := user.(*models.User); ok {
				c.Next()
				return
			}
		}
		seelog.Warnf("User not authorized to visit %s", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}
