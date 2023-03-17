package main

import (
	"dns-search/check"
	"encoding/json"
	"fmt"

	"strings"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	m := make(map[string]string)

	// 生成所有可能的长度为 1 的字母单词
	words := GenerateWords(1)

	// 并行检查所有字母单词的域名是否被注册
	for _, word := range words {
		domain := word + ".com"

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

// GenerateWords 生成所有长度为 n 的字母单词
func GenerateWords(n int) []string {
	letters := make([]byte, n) // 初始化一个长度为 n 的字节数组
	words := []string{}        // 初始化一个空的字符串数组
	generate(letters, 0, &words)
	return words
}

// generate 递归生成所有可能的字母单词
func generate(letters []byte, pos int, words *[]string) {
	if pos == len(letters) { // 如果已经生成了 n 个字母，则将其添加到字符串数组中
		*words = append(*words, string(letters))
		return
	}

	for i := 0; i < 26; i++ { // 生成所有可能的字母，从 a 到 z
		letters[pos] = byte('a' + i)
		generate(letters, pos+1, words) // 递归生成下一个字母
	}
}
