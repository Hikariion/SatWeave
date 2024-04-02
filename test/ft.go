package main

import (
	"fmt"
	"os"
	"satweave/utils/common"
)

func main() {
	//// 假定参数
	//fs := 1000.0 // 采样频率
	//T := 1.0     // 信号时长
	//N := int(T * fs)
	//cutoff := 30.0 // 截止频率
	//
	//for i := 0; i < 10; i++ {
	//	// 直接构建一个模拟的FFT结果数组
	//	signalFFT := make([]complex128, N)
	//	// 选定的频率成分索引和它们的幅度
	//	frequencies := []int{25, 50, 75, 100, 200, 300}
	//	amplitudes := []float64{1.0, 0.8, 0.6, 0.4, 0.2, 0.1}
	//
	//	for _, freq := range frequencies {
	//		// 为选定的频率成分设置不同的幅度和随机相位
	//		phase := rand.Float64() * 2 * math.Pi
	//		amplitude := amplitudes[rand.Intn(len(amplitudes))]
	//		signalFFT[freq] = complex(amplitude*math.Cos(phase), amplitude*math.Sin(phase))
	//
	//		// 考虑FFT结果的对称性
	//		if freq != 0 && freq*2 != N {
	//			signalFFT[N-freq] = complex(amplitude*math.Cos(-phase), amplitude*math.Sin(-phase))
	//		}
	//	}
	//
	//	fmt.Println("应用高通滤波前的非零高频成分数量:", cmplx.Abs(signalFFT[25])+cmplx.Abs(signalFFT[50])+cmplx.Abs(signalFFT[75]))
	//
	//	// 高通滤波
	//	for i := range signalFFT {
	//		freq := float64(i) / T
	//		if freq > fs/2 {
	//			freq -= fs
	//		}
	//		if math.Abs(freq) < cutoff {
	//			signalFFT[i] = 0
	//		}
	//	}
	//
	//	// 计算高频分量的数量
	//	highFreqCount := 0
	//	for _, val := range signalFFT {
	//		if cmplx.Abs(val) > 0 {
	//			highFreqCount++
	//		}
	//	}
	//
	//	fmt.Println("高于 30 Hz 的高频分量数量:", highFreqCount)
	//}
	// 读文件
	file, _ := os.Open("./test-job-id-test-job-id#HaSumOp#-1-1")
	defer file.Close()

	data := make([]byte, 1024)
	file.Read(data)

	res := common.BytesToUint64(data)

	fmt.Print(res)
}
