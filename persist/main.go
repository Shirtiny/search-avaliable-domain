package persist

import (
	"log"
)

func Saver() chan string {

	//输出通道
	out := make(chan string)
	go func() {
		count := 0
		for {
			object := <-out
			//存入es
			//elastic.Add(client, object)
			log.Printf("存储管道输出：#%d %v\n", count, object)
			count++
		}
	}()
	return out
}
