package main

import (
	"dns-search/check"
	"dns-search/engine"
	"dns-search/persist"
	"dns-search/scheduler"
	"encoding/json"
	"fmt"
	"strconv"

	"strings"
	"sync"
)

func main() {
	StartNew()
}

func StartNew() {
	// 生成所有可能的长度为 1 的字母单词
	domains := GenerateDomains(1, ".com")
	tasks := []engine.Task{}

	for i := 0; i < len(domains); i++ {
		tasks = append(tasks, engine.Task{
			Name: "dns-search-" + strconv.Itoa(i),
			Run: func() (engine.TaskResult, error) {
				return engine.TaskResult{
					Data:  []string{domains[i]},
					Tasks: nil,
				}, nil
			},
		})
	}

	engineMain := engine.ConcurrentQueue{
		Scheduler:   &scheduler.QueueScheduler{},
		WorkerCount: 6,
		SaverChan:   persist.Saver(),
	}
	engineMain.Run(engine.Task{
		Name: "dns-search-1",
		Run: func() (engine.TaskResult, error) {
			return engine.TaskResult{
				Data:  []string{"a.com", "b.com"},
				Tasks: tasks,
			}, nil
		},
	})
}

func StartOld() {
	var wg sync.WaitGroup

	m := make(map[string]string)

	// 生成所有可能的长度为 1 的字母单词
	domains := GenerateDomains(1, ".com")

	// 并行检查所有字母单词的域名是否被注册
	for _, domain := range domains {

		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			dnsPass := check.CheckByDNS(d)
			if !dnsPass {
				return
			}
			if available := check.CheckIsDomainAvailableByApi(d); available {
				m[d] = d
			}

		}(domain)
	}

	fmt.Println("waiting Goroutines work done...")

	// 等待所有 Goroutine 执行完毕
	wg.Wait()

	// 创建一个数组，长度为 map 的长度
	results := make([]string, len(m))

	// 遍历 map 的键，并将其存储到数组中
	i := 0
	for k := range m {
		results[i] = m[k]
		i++
	}

	b, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	var result = string(b)

	// 最后如果不要中括号，直接trim掉即可
	result = strings.Trim(result, "[]")

	// 打印
	fmt.Println(result)
}

// GenerateDomains 生成所有长度为 n 的字母单词
func GenerateDomains(n int, suffix string) []string {
	letters := make([]byte, n) // 初始化一个长度为 n 的字节数组
	words := []string{}        // 初始化一个空的字符串数组
	generate(letters, 0, &words, suffix)
	return words
}

// generate 递归生成所有可能的字母单词
func generate(letters []byte, pos int, words *[]string, suffix string) {
	if pos == len(letters) { // 如果已经生成了 n 个字母，则将其添加到字符串数组中
		*words = append(*words, string(letters)+suffix)
		return
	}

	for i := 0; i < 26; i++ { // 生成所有可能的字母，从 a 到 z
		letters[pos] = byte('a' + i)
		generate(letters, pos+1, words, suffix) // 递归生成下一个字母
	}
}
