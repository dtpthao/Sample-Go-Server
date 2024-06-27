package main

import (
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"sample-go-server/entity"
	cH "sample-go-server/module/contract/handler"
	cRepo "sample-go-server/module/contract/repository"
	cUC "sample-go-server/module/contract/usecase"
	tH "sample-go-server/module/token/handler"
	tUC "sample-go-server/module/token/usecase"
	uH "sample-go-server/module/user/handler"
	uRepo "sample-go-server/module/user/repository"
	uUC "sample-go-server/module/user/usecase"
	"sample-go-server/services/kafka"
	apiLog "sample-go-server/utils/log"
	"time"
)

var mysqlConfig entity.MysqlConfig
var serverConfig entity.ServerConfig
var kafkaConfig entity.KafkaConfig

func readConfig() {

	// todo no hardcode
	mysqlConfig = entity.MysqlConfig{
		Host:     "localhost",
		Port:     3307,
		User:     "docker",
		Password: "docker@123",
		DBName:   "dockerdb",
	}

	serverConfig = entity.ServerConfig{
		Host:      "localhost",
		Port:      3030,
		JWTSecret: "5b0b18dc37004b97946367ca5d82673918a6c6e7a817bf84236abe1c0907b9bf",
	}

	kafkaConfig = entity.KafkaConfig{
		BrokerHost:      "localhost",
		BrokerPort:      "9093",
		ConsumerGroupID: "my-group",
	}
}

func SetupDatabase() *gorm.DB {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", mysqlConfig.User, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DBName)
	gormConfig := gorm.Config{}
	gormConfig.DisableNestedTransaction = true
	gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	db, err := gorm.Open(mysql.Open(dns), &gormConfig)
	if err != nil {
		log.Fatalf("Connect to MySQL server failed. %v", err)
	}
	db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Cannot config MySQL connection. %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = db.AutoMigrate(&uRepo.User{}, &cRepo.Contract{})
	if err != nil {
		log.Fatalf("Cannot migrate tables. %v", err)
	}

	return db
}

func main() {

	readConfig()

	db := SetupDatabase()

	jwtSecret, _ := hex.DecodeString(serverConfig.JWTSecret)
	tokenUseCase := tUC.NewTokenUseCase(jwtSecret)

	app := gin.New()
	app.Use(apiLog.AccessLog)

	// user
	userRepository := uRepo.NewUserRepository(db)
	userUseCase := uUC.NewUserUseCase(userRepository, tokenUseCase)

	tokenHandler := tH.NewTokenHandler(tokenUseCase, userUseCase, jwtSecret)

	userHandler := uH.NewUserHandler(app, userUseCase)
	userHandler.RegisterHandler(tokenHandler.Authenticate, tokenHandler.AdminAuthorize)

	kafkaService := kafka.NewKafkaService(kafkaConfig)
	// contract
	cRepository := cRepo.NewContractRepository(db)
	cUseCase := cUC.NewContractUseCase(cRepository)
	cHandler, err := cH.NewContractHandler(app, cUseCase, kafkaService)
	if err != nil {
		panic(err)
	}

	cHandler.RegisterHandler(tokenHandler.Authenticate)

	gin.SetMode(gin.ReleaseMode)
	addr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	log := apiLog.Get()
	log.Info().Msg("Server start at: http://" + addr)
	err = app.Run(addr)

	if err != nil {
		panic(err)
	}

}
