package ioc

import (
	"github.com/jayleonc/geektime-go/webook/internal/repository/dao"
	"github.com/jayleonc/geektime-go/webook/pkg/gormx"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var c Config
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		//Logger: gormlogger.New(log.New(os.Stdout, "\n", log.LstdFlags), gormlogger.Config{
		//	SlowThreshold: time.Second,
		//	Colorful:      true,
		//	LogLevel:      gormlogger.Info,
		//}),
	})
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}
	before := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "gorm_db_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	err = db.Use(before)
	if err != nil {
		panic(err)
	}

	db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithDBName("geektime")))
	dao.InitTables(db)
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{Key: "args", Val: i})
}
