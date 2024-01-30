package operators

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"satweave/utils/errno"
	"satweave/utils/logger"
	"strings"
)

type SimpleSource struct {
	name      string
	wordsChan chan string
	done      chan bool
	counter   uint64
}

func (op *SimpleSource) Init() {
	op.wordsChan = make(chan string, 1)
	op.done = make(chan bool)

	go func() {
		defer close(op.wordsChan)
		file, err := os.Open("./test-files/document.txt")
		if err != nil {
			logger.Errorf("open file failed: %v", err)
			close(op.done)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			word := scanner.Text()
			logger.Infof("read word %s", word)
			if word != "" {
				logger.Infof("send word %s, push", word)
				op.wordsChan <- strings.ToLower(word)
			}
		}
		close(op.done)
	}()
}

func (op *SimpleSource) Compute([]byte) ([]byte, error) {
	select {
	case word, ok := <-op.wordsChan:
		logger.Infof("word: %s", word)
		if !ok {
			// TODO(qiu): 可以返回错误，表示没有单词了
			return nil, errno.JobFinished
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

func (op *SimpleSource) Checkpoint() []byte {
	counter := op.counter
	buf := new(bytes.Buffer)
	// 小端存储
	err := binary.Write(buf, binary.LittleEndian, counter)
	if err != nil {
		logger.Errorf("binary.Write error: %v", err)
		return nil
	}
	return buf.Bytes()
}

func (op *SimpleSource) RestoreFromCheckpoint([]byte) error {
	return nil
}
