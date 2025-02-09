package pkg

func DeepCopyReverseSlice[T any](input []T) []T {
	result := make([]T, len(input))

	for i := 0; i < len(input); i++ {
		result[i] = input[len(input)-1-i]
	}

	return result
}
