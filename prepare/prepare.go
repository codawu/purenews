package prepare

import (
	"github.com/kataras/iris/v12" //nolint:goimports
	"github.com/spf13/pflag"      //nolint:goimports
	"go.uber.org/zap"
	"purenews/config"
	"purenews/log"
	"time" //nolint:goimports
)

var (
	configPath = pflag.StringP("config", "c", "data/config.temp.yaml", "config file path.")
)

type Router interface {
	Init(app *iris.Application)
}

/**
  initial project configuration
*/
func Init() error {
	// parse config file
	pflag.Parse()
	// init config
	if err := config.Init(*configPath); err != nil {
		return err
	}
	if err := log.Init(); err != nil {
		return err
	}
	return nil
}

func listen(router Router) error {
	app := iris.New()
	router.Init(app)
	address := config.Config.Server.Address
	if err := app.Listen(address, iris.WithoutServerError(iris.ErrServerClosed), iris.WithConfiguration(iris.Configuration{
		DisableInterruptHandler:   false,
		DisablePathCorrection:     false,
		EnablePathEscape:          false,
		FireMethodNotAllowed:      false,
		DisableAutoFireStatusCode: false,
		TimeFormat:                "Mon, 02 Jan 2006 15:04:05 GMT",
		Charset:                   "UTF-8",
	})); err != nil {
		return err
	}
	return nil
}

/**
  start web service
*/
func Serve(router Router) error {
	return listen(router)
}

type Task interface {
	Run() (interface{}, error)
	GetRetry() int
	SetRetry(int)
}

type TaskQueue interface {
	Push(task Task) error
	Pop() (Task, error)
}

//
func Worker(tick *<-chan time.Time, queue TaskQueue) error {
	for {
		if tick != nil {
			<-*tick
		}
		task, err := queue.Pop()
		if err != nil {
			log.Log().Error("QueuePop", zap.Any("queue", queue), zap.Error(err))
			return err
		}
		log.Log().Info("QueuePop", zap.Any("queue", queue), zap.Any("task", task), zap.Error(err))
		result, err := task.Run()
		if err != nil {
			log.Log().Error("RunTask", zap.Any("task", task), zap.Error(err))
			if task.GetRetry() > 0 {
				task.SetRetry(task.GetRetry() - 1)
				go func(task Task) {
					if err := queue.Push(task); err != nil {
						log.Log().Error("QueuePush", zap.Any("queue", queue), zap.Any("task", task), zap.Error(err))
					} else {
						log.Log().Info("QueuePush", zap.Any("queue", queue), zap.Any("task", task), zap.Error(err))
					}
				}(task)
			}
			continue
		}
		log.Log().Info("RunTask", zap.Any("result", result))
	}
}
