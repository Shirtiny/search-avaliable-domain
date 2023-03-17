package engine

import (
	"fmt"
	"log"
	"sync"
)

// Task，存储url、解析器方法
type Task struct {
	Name string
	Run  RunFunc
}

type TaskResult struct {
	Data  []string
	Tasks []Task
}

type RunFunc func() (TaskResult, error)

// 队列版本 并发引擎 正在用
type ConcurrentQueue struct {
	//调度器
	Scheduler SchedulerQueue
	//worker协程数
	WorkerCount int
	//存储通道 传输
	SaverChan chan string

	allDone bool
}

// 接口 调度器 接口方法的参数不需要名字
type SchedulerQueue interface {
	//将task输入worker
	SubmitQueue(Task)
	//接收worker的已就绪通知 接收通知后会将传入的workerIn输入通道交给调度器要调度的workerIn管道
	WorkerAlreadyQueue(chan Task)
	//使用协程构建tasks队列和worker队列
	RunQueue(workerTotalCount int, engineDone func())
}

// 将所有task作为参数依次传入seeds数组中
func (engine *ConcurrentQueue) Run(wg *sync.WaitGroup, tasks ...Task) {

	//运行队列调度器 生成task队列和workIn队列 等待task和workerIn
	engine.Scheduler.RunQueue(engine.WorkerCount, engine.Done)

	//out负责worker输出 传输TaskResult
	out := make(chan TaskResult)

	//创建workerCount个worker协程
	for i := 0; i < engine.WorkerCount; i++ {
		createWorkerQueue(engine.Scheduler, out)
	}

	//将所有task交给调度器 调度器对worker输入
	for _, task := range tasks {
		engine.Scheduler.SubmitQueue(task)
	}

	count := 0

	//处理worker的输出结果
	for {
		if engine.allDone {
			fmt.Println("全部任务完成")
			engine.SendToSaver("done")
			wg.Wait()
			break
		}

		//接收输出结果
		result := <-out

		//打印解析结果中的objects result.Data
		for _, data := range result.Data {
			engine.SendToSaver(data)

		}

		//处理解析出的新请求 result.Tasks
		for _, task := range result.Tasks {
			//将新请求提交给调度器处理 => worker输出中取task输入到scheduler
			engine.Scheduler.SubmitQueue(task)
		}

		count++

		// fmt.Println("完成的任务数：", count)
	}

}

// 关闭引擎
func (engine *ConcurrentQueue) Done() {
	engine.allDone = true
}

//传入存储通道
func (engine *ConcurrentQueue) SendToSaver(data string) {
	go func(d string) {
		log.Printf("存储管道输入： %v\n", d)
		engine.SaverChan <- d
	}(data)
}

// 整合了fetcher和Parser
func working(task Task) (TaskResult, error) {
	//使用fetcher发起请求，得到请求url后的数据
	result, err := task.Run()
	// fmt.Printf("task的Name：%s\n", task.Name)
	if err != nil {
		//处理错误err
		log.Printf("task执行失败,url：%s ；error：%v", task.Name, err)

		return TaskResult{}, err
	}
	//解析fetcher请求url返回的数据
	return result, nil
}

// 创建worker协程
func createWorkerQueue(scheduler SchedulerQueue, out chan TaskResult) {

	//每个worker都有一个自己的in channel
	in := make(chan Task)

	//协程
	go func() {
		for {
			//告诉调度器 此worker已经就绪 把in传过去
			scheduler.WorkerAlreadyQueue(in)
			//从task管道中取出task请求
			task := <-in
			//调用worker处理请求
			result, e := working(task)
			//出错 结束本轮循环，进入下一轮 不对result处理
			if e != nil {
				continue
			}
			//没出错则将result放入ParseResult管道
			out <- result
		}
	}()
}
