package utils

import (
	"fmt"
	"math"
	"strconv"
)

// ToInt chuyển string sang int
func ToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// ToInt64 chuyển string sang int64
func ToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// ToFloat64 chuyển string sang float64
func ToFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ToString chuyển int sang string
func ToString(i int) string {
	return strconv.Itoa(i)
}

// Round làm tròn số thập phân
func Round(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}

// RoundUp làm tròn lên
func RoundUp(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Ceil(num*output) / output
}

// RoundDown làm tròn xuống
func RoundDown(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Floor(num*output) / output
}

// FormatMoney format số tiền (1000000 -> 1,000,000)
func FormatMoney(amount float64) string {
	s := fmt.Sprintf("%.0f", amount)

	// Add commas
	n := len(s)
	if n <= 3 {
		return s
	}

	result := ""
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}

	return result
}

// FormatMoneyVND format số tiền VND (1000000 -> 1.000.000đ)
func FormatMoneyVND(amount float64) string {
	s := fmt.Sprintf("%.0f", amount)

	// Add dots
	n := len(s)
	if n <= 3 {
		return s + "đ"
	}

	result := ""
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	return result + "đ"
}

// Percentage tính phần trăm
func Percentage(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (part / total) * 100
}

// PercentageChange tính % thay đổi
func PercentageChange(oldValue, newValue float64) float64 {
	if oldValue == 0 {
		return 0
	}
	return ((newValue - oldValue) / oldValue) * 100
}

// InRange kiểm tra số có trong khoảng không
func InRange(num, min, max float64) bool {
	return num >= min && num <= max
}

// Clamp giới hạn số trong khoảng
func Clamp(num, min, max float64) float64 {
	if num < min {
		return min
	}
	if num > max {
		return max
	}
	return num
}

// Min trả về số nhỏ nhất
func Min(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	min := numbers[0]
	for _, num := range numbers[1:] {
		if num < min {
			min = num
		}
	}

	return min
}

// Max trả về số lớn nhất
func Max(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	max := numbers[0]
	for _, num := range numbers[1:] {
		if num > max {
			max = num
		}
	}

	return max
}

// Sum tính tổng
func Sum(numbers ...float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

// Average tính trung bình
func Average(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return Sum(numbers...) / float64(len(numbers))
}

// MinInt trả về int nhỏ nhất
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt trả về int lớn nhất
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AbsInt trả về giá trị tuyệt đối
func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AbsFloat64 trả về giá trị tuyệt đối
func AbsFloat64(n float64) float64 {
	return math.Abs(n)
}
