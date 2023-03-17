package engine

import (
	"fmt"
	"log"
)

// Request，存储url、解析器方法
type Request struct {
	Url        string
	ParserFunc ParserFunc
}

type ParseResult struct {
	Objects []string
}

type ParserFunc func([]byte) ParseResult

// 队列版本 并发引擎 正在用
type ConcurrentQueue struct {
	//调度器
	Scheduler SchedulerQueue
	//worker协程数
	WorkerCount int
	//存储通道 传输
	SaverChan chan string
}

// 接口 调度器 接口方法的参数不需要名字
type SchedulerQueue interface {
	//将request输入worker
	SubmitQueue(Request)
	//接收worker的已就绪通知 接收通知后会将传入的workerIn输入通道交给调度器要调度的workerIn管道
	WorkerAlreadyQueue(chan Request)
	//使用协程构建requests队列和worker队列
	RunQueue()
}

// 将所有request作为参数依次传入seeds数组中
func (engine *ConcurrentQueue) Run(seeds ...Request) {

	//运行队列调度器 生成request队列和workIn队列 等待request和workerIn
	engine.Scheduler.RunQueue()

	//out负责worker输出 传输ParseResult
	out := make(chan ParseResult)

	//创建workerCount个worker协程
	for i := 0; i < engine.WorkerCount; i++ {
		createWorkerQueue(engine.Scheduler, out)
	}

	//将所有request交给调度器 调度器对worker输入
	for _, request := range seeds {
		engine.Scheduler.SubmitQueue(request)
	}

	//处理worker的输出结果
	for {
		//接收输出结果
		result := <-out

		//打印解析结果中的objects result.Objects
		for _, object := range result.Objects {
			//传入存储通道
			go func(esObject string) {
				log.Printf("存储管道输入： %v\n", esObject)
				engine.SaverChan <- esObject
			}(object)
		}

		//处理解析出的新请求 result.Requests
		for _, request := range result.Requests {
			//将新请求提交给调度器处理 => worker输出中取request输入到scheduler
			engine.Scheduler.SubmitQueue(request)
		}

	}

}

// 整合了fetcher和Parser
func workerQueue(request Request) (ParseResult, error) {
	//使用fetcher发起请求，得到请求url后的数据
	bytes, err := fetcher.Fetcher(request.Url)
	fmt.Printf("请求的URL：%s\n", request.Url)
	if err != nil {
		//处理错误err
		log.Printf("fetcher请求失败,url：%s ；error：%v", request.Url, err)

		return ParseResult{}, err
	}
	//解析fetcher请求url返回的数据
	return request.ParserFunc(bytes), nil
}

// 创建worker协程
func createWorkerQueue(scheduler SchedulerQueue, out chan ParseResult) {

	//每个worker都有一个自己的in channel
	in := make(chan Request)

	//协程
	go func() {
		for {
			//告诉调度器 此worker已经就绪 把in传过去
			scheduler.WorkerAlreadyQueue(in)
			//从request管道中取出request请求
			request := <-in
			//调用worker处理请求
			result, e := workerQueue(request)
			//出错 结束本轮循环，进入下一轮 不对result处理
			if e != nil {
				continue
			}
			//没出错则将result放入ParseResult管道
			out <- result
		}
	}()
}
