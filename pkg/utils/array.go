package utils

// UniqueStrings loại bỏ duplicate trong slice string
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// UniqueInts loại bỏ duplicate trong slice int
func UniqueInts(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// FilterStrings lọc slice string theo condition
func FilterStrings(slice []string, fn func(string) bool) []string {
	result := []string{}
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterInts lọc slice int theo condition
func FilterInts(slice []int, fn func(int) bool) []int {
	result := []int{}
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapStrings áp dụng function cho mỗi phần tử
func MapStrings(slice []string, fn func(string) string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// MapInts áp dụng function cho mỗi phần tử
func MapInts(slice []int, fn func(int) int) []int {
	result := make([]int, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// ChunkStrings chia slice thành các chunks
func ChunkStrings(slice []string, size int) [][]string {
	if size <= 0 {
		return nil
	}

	chunks := [][]string{}
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ChunkInts chia slice thành các chunks
func ChunkInts(slice []int, size int) [][]int {
	if size <= 0 {
		return nil
	}

	chunks := [][]int{}
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ReverseStrings đảo ngược slice
func ReverseStrings(slice []string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		result[len(slice)-1-i] = item
	}
	return result
}

// ReverseInts đảo ngược slice
func ReverseInts(slice []int) []int {
	result := make([]int, len(slice))
	for i, item := range slice {
		result[len(slice)-1-i] = item
	}
	return result
}

// ContainsInt kiểm tra int có trong slice không
func ContainsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IndexOf tìm index của item trong slice
func IndexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// IndexOfInt tìm index của item trong slice int
func IndexOfInt(slice []int, item int) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// RemoveString xóa item khỏi slice
func RemoveString(slice []string, item string) []string {
	result := []string{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// RemoveInt xóa item khỏi slice
func RemoveInt(slice []int, item int) []int {
	result := []int{}
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// DifferenceStrings trả về items có trong slice1 nhưng không có trong slice2
func DifferenceStrings(slice1, slice2 []string) []string {
	result := []string{}
	set := make(map[string]bool)

	for _, item := range slice2 {
		set[item] = true
	}

	for _, item := range slice1 {
		if !set[item] {
			result = append(result, item)
		}
	}

	return result
}

// IntersectionStrings trả về items có trong cả 2 slices
func IntersectionStrings(slice1, slice2 []string) []string {
	result := []string{}
	set := make(map[string]bool)

	for _, item := range slice1 {
		set[item] = true
	}

	for _, item := range slice2 {
		if set[item] {
			result = append(result, item)
			delete(set, item) // Tránh duplicate
		}
	}

	return result
}

// UnionStrings hợp 2 slices (không duplicate)
func UnionStrings(slice1, slice2 []string) []string {
	set := make(map[string]bool)
	result := []string{}

	for _, item := range slice1 {
		if !set[item] {
			set[item] = true
			result = append(result, item)
		}
	}

	for _, item := range slice2 {
		if !set[item] {
			set[item] = true
			result = append(result, item)
		}
	}

	return result
}
