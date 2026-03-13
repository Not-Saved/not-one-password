package utils

import "strconv"

func Int32FromString(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(i), nil
}
