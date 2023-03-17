package persist

import (
	"encoding/json"
	"log"
	"sync"
)

func Saver(wg *sync.WaitGroup) (saverChan chan string) {

	m := make(map[string]string)

	//输出通道
	out := make(chan string)

	wg.Add(1)

	go func(mp map[string]string) {

		for count := 0; ; count++ {

			// object这里简化为string 可传递更复杂的struct
			object := <-out
			log.Printf("存储管道输出：#%d %v\n", count, object)

			// 以done作为存储结束标志
			if object == "done" {
				results := make([]string, len(mp))
				i := 0
				for k := range mp {
					results[i] = k
					i++
				}
				b, _ := json.Marshal(results)
				log.Println("total: ", string(b))
				wg.Done()
				break
			}

			mp[object] = ""

		}
	}(m)

	return out
}
