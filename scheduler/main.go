package scheduler

import "dns-search/engine"

//队列版本调度器
type QueueScheduler struct {
	//用于传输Request的通道
	requestChannel chan engine.Task
	//workers的每个worker的输入通道 这些子输入管道 将共用一个总管道
	workerInChannel chan chan engine.Task
}

//提交请求
func (scheduler *QueueScheduler) SubmitQueue(request engine.Task) {
	scheduler.requestChannel <- request
}

//接收worker的已就绪通知 接收通知后会将传入的workerIn输入通道交给调度器要调度的workerIn管道
func (scheduler *QueueScheduler) WorkerAlreadyQueue(workerIn chan engine.Task) {
	scheduler.workerInChannel <- workerIn
}

//使用协程构建requests队列和worker队列
func (scheduler *QueueScheduler) RunQueue() {

	//生成需要的channel
	scheduler.requestChannel = make(chan engine.Task)
	scheduler.workerInChannel = make(chan chan engine.Task)

	go func() {
		//request队列
		var requests []engine.Task
		//workerIn队列
		var workerIns []chan engine.Task
		for {
			//活跃状态的Request
			var activeRequest engine.Task
			//就绪状态的workerIn
			var alreadyWorkerIn chan engine.Task
			//从队列中读取第一个request和workerIn
			if len(requests) > 0 && len(workerIns) > 0 {
				activeRequest = requests[0]
				alreadyWorkerIn = workerIns[0]
			}

			//select的含义去看看教学
			select {
			//独立的收取request和workerIn 放入对应的队列
			case request := <-scheduler.requestChannel:
				requests = append(requests, request)

			case workerIn := <-scheduler.workerInChannel:
				workerIns = append(workerIns, workerIn)

			//把要处理的请求送给就绪的workerIn 然后把干活去的request和worker从队列中去除
			case alreadyWorkerIn <- activeRequest:
				requests = requests[1:]
				workerIns = workerIns[1:]
			}
		}
	}()
}
