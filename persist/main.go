package persist

import (
	"encoding/json"
	"log"
	"strings"
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
				var sb strings.Builder
				i := 0
				for k := range mp {
					results[i] = k
					sb.WriteString(k + ",")
					i++
				}
				b, _ := json.Marshal(results)

				str := sb.String()

				log.Println("json: ", string(b))
				log.Println("total: ", str[:len(str)-1])
				wg.Done()
				break
			}

			mp[object] = ""

		}
	}(m)

	return out
}
