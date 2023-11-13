package operators

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type SimpleSource struct {
	name      string
	wordsChan chan string
	done      chan bool
}

func (op *SimpleSource) Init(map[string]string) {
	op.wordsChan = make(chan string, 10)
	op.done = make(chan bool)

	go func() {
		defer close(op.wordsChan)
		file, err := os.Open("./document")
		if err != nil {
			close(op.done)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		re := regexp.MustCompile("[.,\\s]+")

		for scanner.Scan() {
			line := scanner.Text()
			words := re.Split(line, -1)
			for _, word := range words {
				if word != "" {
					op.wordsChan <- strings.ToLower(word)
				}
			}
		}
		close(op.done)
	}()
}

func (op *SimpleSource) Compute([]byte) ([]byte, error) {
	select {
	case word, ok := <-op.wordsChan:
		if !ok {
			// TODO(qiu): 可以返回错误，表示没有单词了
			return nil, nil
		}
		return []byte(word), nil
	case <-op.done:
		// TODO(qiu): 可以返回错误，表示没有单词了
		return nil, nil
	}
}

func (op *SimpleSource) SetName(name string) {
	op.name = name
}

func (op *SimpleSource) IsSourceOp() bool {
	return true
}

func (op *SimpleSource) IsSinkOp() bool {
	return false
}

func (op *SimpleSource) IsKeyByOp() bool {
	return false
}
