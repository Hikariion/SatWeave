package main

import (
	"fmt"
	"math"
	"math/cmplx"
)

import "github.com/mjibson/go-dsp/fft"

func main() {
	// 定义参数
	fs := 1000.0   // 采样频率
	T := 1.0       // 信号时长
	fLow := 5.0    // 低频分量
	fHigh := 50.0  // 高频分量
	cutoff := 30.0 // 截止频率

	// 生成模拟信号
	N := int(T * fs) // 样本数量

	signal := make([]complex128, N)

	for n := 0; n < N; n++ {
		t := float64(n) / fs
		signal[n] = complex(math.Sin(2*math.Pi*fLow*t)+0.5*math.Sin(2*math.Pi*fHigh*t), 0)
	}

	// FFT
	signalFFT := fft.FFT(signal)

	// 高通滤波
	for i, _ := range signalFFT {
		freq := float64(i) / T
		if freq > fs/2 {
			freq = freq - fs
		}
		if math.Abs(freq) < cutoff {
			signalFFT[i] = 0
		}
	}

	// 计算高频分量的数量
	highFreqCount := 0
	for _, val := range signalFFT {
		if cmplx.Abs(val) > 0 {
			highFreqCount++
		}
	}

	fmt.Printf("高于 %v Hz 的高频分量数量: %v\n", cutoff, highFreqCount)
}
