package persist

import (
	"encoding/json"
	"log"
)

func Saver() chan string {

	m := make(map[string]string)

	//输出通道
	out := make(chan string)

	go func(mp map[string]string) {

		for count := 0; ; count++ {
			object := <-out
			//存入es
			//elastic.Add(client, object)
			log.Printf("存储管道输出：#%d %v\n", count, object)
			mp[object] = object
			b, _ := json.Marshal(mp)
			log.Println("total: ", string(b))
			log.Println("waiting next...")
		}
	}(m)

	return out
}
