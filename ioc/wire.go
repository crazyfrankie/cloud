//go:build wireinject

package ioc

import (
	"fmt"
	"os"

	"github.com/crazyfrankie/cloud/internal/auth"
	"github.com/crazyfrankie/cloud/internal/file"
	"github.com/crazyfrankie/cloud/internal/storage"
	"github.com/crazyfrankie/cloud/pkg/conf"
	"github.com/crazyfrankie/cloud/pkg/middlewares"
	snowflake "github.com/crazyfrankie/snow-flake"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	fdao "github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/user"
	udao "github.com/crazyfrankie/cloud/internal/user/dao"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(conf.GetConf().MySQL.DSN,
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&udao.User{}, &fdao.File{})

	return db
}

func InitMinIO() *minio.Client {
	client, err := minio.New(conf.GetConf().MinIO.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(conf.GetConf().MinIO.AccessKey, conf.GetConf().MinIO.SecretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	return client
}

func InitRedis() redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr: conf.GetConf().Redis.Addr,
	})

	return client
}

func InitSnowflake() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	return node
}

func InitWeb(mws []gin.HandlerFunc, user *user.Handler, auth *auth.Handler,
	storage *storage.Handler, file *file.Handler) *gin.Engine {
	srv := gin.Default()

	srv.Use(mws...)

	user.RegisterRoute(srv)
	auth.RegisterRoute(srv)
	storage.RegisterRoute(srv)
	file.RegisterRoute(srv)

	return srv
}

func InitMws(t *auth.TokenService) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "x-access-token")
			c.Header("Access-Control-Allow-Credentials", "true")

			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}

			c.Next()
		},

		middlewares.NewAuthnHandler(t).
			IgnorePath("/user/register").
			IgnorePath("/auth/login").
			Auth(),
	}
}

func InitEngine() *gin.Engine {
	wire.Build(
		InitDB,
		InitRedis,
		InitMinIO,
		InitSnowflake,

		user.InitUserModule,
		auth.InitAuthModule,
		storage.InitStorageModule,
		file.InitFileModule,

		InitMws,
		InitWeb,

		wire.FieldsOf(new(*user.Module), "Handler"),
		wire.FieldsOf(new(*auth.Module), "Handler"),
		wire.FieldsOf(new(*auth.Module), "Token"),
		wire.FieldsOf(new(*storage.Module), "Handler"),
		wire.FieldsOf(new(*file.Module), "Handler"),
	)
	return new(gin.Engine)
}
