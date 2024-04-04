package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	//const Round = 1000
	//// 假定参数
	//fs := 1000.0 // 采样频率
	//T := 1.0     // 信号时长
	//N := int(T * fs)
	//
	//var signals [][]complex128
	//
	//for i := 0; i < Round; i++ {
	//	rand.Seed(time.Now().UnixNano())
	//
	//	signal := make([]complex128, N)
	//	fLow := rand.Float64()*5 + 5              // 产生5到10之间的随机低频
	//	fHigh := rand.Float64()*45 + 5            // 产生5到50之间的随机高频
	//	phaseLow := rand.Float64() * 2 * math.Pi  // 为低频生成随机相位
	//	phaseHigh := rand.Float64() * 2 * math.Pi // 为高频生成随机相位
	//	for n := 0; n < N; n++ {
	//		if rand.Float64() <= 2.0/3.0 {
	//			t := float64(n) / fs
	//			signal[n] = complex(math.Sin(2*math.Pi*fLow*t+phaseLow)+0.5*math.Sin(2*math.Pi*fHigh*t+phaseHigh), 0)
	//		}
	//	}
	//
	//	signals = append(signals, signal)
	//}
	//
	//// 将信号写入文件
	//file, err := os.Create("signals.txt")
	//if err != nil {
	//	fmt.Println("Error creating file:", err)
	//	return
	//}
	//defer file.Close()
	//
	//for i, signal := range signals {
	//	fmt.Fprintf(file, "%d\n", i)
	//	for _, value := range signal {
	//		// 为了简化，这里只写入实部和虚部
	//		fmt.Fprintf(file, " %f, %f", real(value), imag(value))
	//	}
	//	fmt.Fprintln(file) // 每个信号后换行
	//}
	//fmt.Println("Signals have been written to signals.txt")
	file, _ := os.Open("signals.txt")
	defer file.Close()
	signal, _ := readSignalByID(file, 0)
	fmt.Println(signal)
	signal, _ = readSignalByID(file, 1)
	fmt.Println(signal)
}

func readSignalByID(file *os.File, id int) ([]complex128, error) {
	scanner := bufio.NewScanner(file)
	var foundID bool
	var signal []complex128

	for scanner.Scan() {
		line := scanner.Text()

		// 尝试将行转换为ID
		currentID, err := strconv.Atoi(line)
		if err == nil {
			if currentID == id {
				foundID = true // 找到匹配的ID
				continue       // 继续读取下一行以获取信号数据
			}
		}

		if foundID {
			// 找到ID后，当前行包含信号数据
			dataParts := strings.Split(line, ", ")
			for _, part := range dataParts {
				values := strings.Split(part, " ")
				if len(values) != 2 {
					continue // 如果数据格式不符，跳过此数据点
				}
				realPart, err := strconv.ParseFloat(values[0], 64)
				if err != nil {
					continue // 如果实部无法转换，跳过此数据点
				}
				imagPart, err := strconv.ParseFloat(values[1], 64)
				if err != nil {
					continue // 如果虚部无法转换，跳过此数据点
				}
				signal = append(signal, complex(realPart, imagPart))
			}
			return signal, nil // 返回找到的信号
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if !foundID {
		return nil, fmt.Errorf("signal not found for ID %d", id)
	}
	return signal, nil
}
