package persist

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

				chunks := toChunk(results, 20)
				mkdir("./temp")

				for i, chunk := range chunks {
					var tsb strings.Builder
					for _, domain := range chunk {
						tsb.WriteString(domain + "\n")
					}

					str := tsb.String()
					writeFile("./temp/", strconv.Itoa(i)+".txt", str[:len(str)-1])

				}

				wg.Done()
				break
			}

			mp[object] = ""

		}
	}(m)

	return out
}

func toChunk(arr []string, chunkSize int) [][]string {
	var chunks [][]string

	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize

		if end > len(arr) {
			end = len(arr)
		}

		chunks = append(chunks, arr[i:end])
	}

	return chunks
}

func mkdir(dir string) {
	// 检查是否存在 temp 目录
	if _, err := os.Stat(dir); err == nil {
		// 如果存在，删除 temp 目录及其所有内容
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println("Error deleting temp directory:", err)
			return
		}
	}

	// 创建 temp 目录
	err := os.Mkdir(dir, 0755)
	if err != nil {
		fmt.Println("Error creating temp directory:", err)
		return
	}

}

func writeFile(dir string, fileName string, data string) {
	file, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Data written to file successfully!")
}
