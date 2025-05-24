package utils

func GetStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func GetInt64Value(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}