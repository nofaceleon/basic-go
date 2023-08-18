package main

import (
	"gitee.com/geekbang/basic-go/webook/internal/config"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()
	u := initUser(db)
	u.RegisterRoutes(server)
	//server := gin.Default()
	server.Run(":8080")
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: "",
		DB:       0,
	})
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	//处理跨域
	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		//ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello world")
		return
	})

	// 步骤1
	//store := cookie.NewStore([]byte("secret"))
	//server.Use(sessions.Sessions("mysession", store))
	// 步骤3

	//这边是进行登录校验的地方
	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signup"). //拼接不需要进行登录校验的地方
		IgnorePaths("/users/login").
		IgnorePaths("/users/loginJwt").
		Build())
	// v1
	//middleware.IgnorePaths = []string{"sss"}
	//server.Use(middleware.CheckLogin())

	// 不能忽略sss这条路径
	//server1 := gin.Default()
	//server1.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	//db.Debug() //开启debug模式
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	//使用mysql
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	//db, err := gorm.Open(sqlite.Open("webook.db"), &gorm.Config{})
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
