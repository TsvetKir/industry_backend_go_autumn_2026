package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	var result Stats
	if len(nums) < 2 {
		return result
	}

	firstDiff := nums[1] - nums[0]
	result.Min = firstDiff
	result.Max = firstDiff
	result.Sum = firstDiff
	result.Count = 1

	for i := 2; i < len(nums); i++ {
		diff := nums[i] - nums[i-1]
		result.Count += 1
		result.Sum += diff
		if result.Max < diff {
			result.Max = diff
		}
		if result.Min > diff {
			result.Min = diff
		}
	}

	return result
}
